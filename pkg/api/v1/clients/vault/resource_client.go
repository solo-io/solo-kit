package vault

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/solo-io/solo-kit/pkg/api/shared"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/utils/protoutils"
	"k8s.io/apimachinery/pkg/labels"

	vault "github.com/hashicorp/vault/api"
	auth "github.com/hashicorp/vault/api/auth/userpass"
)

const (
	dataKey        = "data"
	optionsKey     = "options"
	checkAndSetKey = "cas"
)

func (rc *ResourceClient) fromVaultSecret(secret *vault.Secret) (resources.Resource, bool, error) {
	if secret.Data == nil {
		return nil, false, errors.Errorf("secret data cannot be nil")
	}
	data, err := parseDataResponse(secret.Data)
	if err != nil {
		return nil, false, errors.Wrapf(err, "parsing data response")
	}
	// if deletion time set, the secret was deleted
	deleted := data.Metadata.DeletionTime != "" || data.Metadata.Destroyed

	resource := rc.NewResource()
	if err := protoutils.UnmarshalMap(data.Data, resource); err != nil {
		return nil, false, err
	}
	resources.UpdateMetadata(resource, func(meta *core.Metadata) {
		meta.ResourceVersion = strconv.Itoa(data.Metadata.Version)
	})
	return resource, deleted, nil
}

func (rc *ResourceClient) toVaultSecret(resource resources.Resource) (map[string]interface{}, error) {
	var version int
	if rv := resource.GetMetadata().ResourceVersion; rv != "" {
		var err error
		version, err = strconv.Atoi(resource.GetMetadata().ResourceVersion)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid resource version: %v (must be int)", rv)
		}
	} else {
		version = 0
	}

	values := make(map[string]interface{})
	data, err := protoutils.MarshalMap(resource)
	if err != nil {
		return nil, err
	}
	values[dataKey] = data
	values[optionsKey] = map[string]interface{}{
		checkAndSetKey: version,
	}
	return values, nil
}

type ResourceClient struct {
	vault *vault.Client

	// Vault's path where resources are located.
	root string

	// Tells Vault which secrets engine it should route traffic to. Defaults to "secret".
	// https://learn.hashicorp.com/tutorials/vault/getting-started-secrets-engines
	pathPrefix   string
	resourceType resources.VersionedResource
}

func NewResourceClient(client *vault.Client, pathPrefix string, rootKey string, resourceType resources.VersionedResource) *ResourceClient {
	if pathPrefix == "" {
		pathPrefix = "secret"
	}

	return &ResourceClient{
		vault:        client,
		pathPrefix:   pathPrefix,
		root:         rootKey,
		resourceType: resourceType,
	}
}

func (rc *ResourceClient) Kind() string {
	return resources.Kind(rc.resourceType)
}

func (rc *ResourceClient) NewResource() resources.Resource {
	return resources.Clone(rc.resourceType)
}

func (rc *ResourceClient) Register() error {
	return nil
}

func (rc *ResourceClient) Read(namespace, name string, opts clients.ReadOpts) (resources.Resource, error) {
	if err := resources.ValidateName(name); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	opts = opts.WithDefaults()
	key := rc.resourceKey(namespace, name)

	secret, err := rc.vault.Logical().Read(key)
	if err != nil {
		return nil, errors.Wrapf(err, "performing vault KV get")
	}
	if secret == nil {
		return nil, errors.NewNotExistErr(namespace, name)
	}

	resource, deleted, err := rc.fromVaultSecret(secret)
	if err != nil {
		return nil, err
	}
	if deleted {
		return nil, errors.NewNotExistErr(namespace, name)
	}
	return resource, nil
}

func (rc *ResourceClient) Write(resource resources.Resource, opts clients.WriteOpts) (resources.Resource, error) {
	opts = opts.WithDefaults()
	if err := resources.ValidateName(resource.GetMetadata().Name); err != nil {
		return nil, errors.Wrapf(err, "validation error")
	}
	meta := resource.GetMetadata()

	if meta.Namespace == "" {
		return nil, errors.Errorf("namespace cannot be empty for vault-backed resources")
	}
	key := rc.resourceKey(meta.Namespace, meta.Name)

	original, err := rc.Read(meta.Namespace, meta.Name, clients.ReadOpts{})
	if original != nil && err == nil {
		if !opts.OverwriteExisting {
			return nil, errors.NewExistErr(meta)
		}
	}

	// mutate and return clone
	clone := resources.Clone(resource)
	clone.SetMetadata(meta)

	secret, err := rc.toVaultSecret(clone)
	if err != nil {
		return nil, err
	}

	if _, err := rc.vault.Logical().Write(key, secret); err != nil {
		return nil, errors.Wrapf(err, "writing to KV")
	}
	// return a read object to update the modify index
	return rc.Read(meta.Namespace, meta.Name, clients.ReadOpts{Ctx: opts.Ctx})
}

func (rc *ResourceClient) ApplyStatus(statusClient resources.StatusClient, inputResource resources.InputResource, opts clients.ApplyStatusOpts) (resources.Resource, error) {
	return shared.ApplyStatus(rc, statusClient, inputResource, opts)
}

func (rc *ResourceClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	opts = opts.WithDefaults()

	if namespace == "" {
		return errors.Errorf("namespace cannot be empty for vault-backed resources")
	}

	if !opts.IgnoreNotExist {
		if _, err := rc.Read(namespace, name, clients.ReadOpts{Ctx: opts.Ctx}); err != nil {
			return err
		}
	}
	metaKey := rc.resourceMetadataKey(namespace, name)

	if _, err := rc.vault.Logical().Delete(metaKey); err != nil {
		return errors.Wrapf(err, "deleting resource %v", name)
	}
	return nil
}

func (rc *ResourceClient) List(namespace string, opts clients.ListOpts) (resources.ResourceList, error) {
	if namespace != "" {
		// list on a single namespace
		return rc.listSingleNamespace(namespace, opts)
	}

	// handle NamespaceAll case

	var namespaces []string

	resourceMetaDir := rc.resourceDirectory("", directoryTypeMetadata)

	namespaces, err := rc.listKeys(resourceMetaDir)
	if err != nil {
		return nil, errors.Wrapf(err, "reading namespace root")
	}

	var resourceList resources.ResourceList
	for _, ns := range namespaces {
		nsResources, err := rc.listSingleNamespace(ns, opts)
		if err != nil {
			return nil, err
		}
		resourceList = append(resourceList, nsResources...)
	}
	return resourceList.Sort(), nil
}

func (rc *ResourceClient) listSingleNamespace(namespace string, opts clients.ListOpts) (resources.ResourceList, error) {
	opts = opts.WithDefaults()

	resourceMetaDir := rc.resourceDirectory(namespace, directoryTypeMetadata)

	secretKeys, err := rc.listKeys(resourceMetaDir)
	if err != nil {
		return nil, errors.Wrapf(err, "reading resource namespace directory")
	}

	var resourceList resources.ResourceList
	for _, key := range secretKeys {
		secret, err := rc.vault.Logical().Read(rc.resourceDirectory(namespace, directoryTypeData) + "/" + key)
		if err != nil {
			return nil, errors.Wrapf(err, "getting secret %s", key)
		}
		if secret == nil {
			return nil, errors.Errorf("unexpected nil err on %v", key)
		}

		resource, deleted, err := rc.fromVaultSecret(secret)
		if err != nil {
			return nil, err
		}
		if !deleted && labels.SelectorFromSet(opts.Selector).Matches(labels.Set(resource.GetMetadata().Labels)) {
			resourceList = append(resourceList, resource)
		}
	}
	return resourceList.Sort(), nil
}

// list on a single namespace
func (rc *ResourceClient) listKeys(directory string) ([]string, error) {
	keyList, err := rc.vault.Logical().List(directory)
	if err != nil {
		renewToken(rc.vault)
		return nil, errors.Wrapf(err, "listing directory %v", directory)
	}
	if keyList == nil {
		return []string{}, nil
	}
	val, ok := keyList.Data["keys"]
	if !ok {
		return nil, errors.Errorf("vault secret list at root %s did not contain key \"keys\"", directory)
	}
	keys, ok := val.([]interface{})
	if !ok {
		return nil, errors.Errorf("expected secret list of type []interface{} but got %v", reflect.TypeOf(val))
	}

	var keysAsStrings []string
	for _, keyAsInterface := range keys {
		key, ok := keyAsInterface.(string)
		if !ok {
			return nil, errors.Errorf("expected key of type string but got %v", reflect.TypeOf(keyAsInterface))
		}
		keysAsStrings = append(keysAsStrings, key)
	}

	return keysAsStrings, nil
}

func (rc *ResourceClient) Watch(namespace string, opts clients.WatchOpts) (<-chan resources.ResourceList, <-chan error, error) {
	opts = opts.WithDefaults()

	resourcesChan := make(chan resources.ResourceList)
	errs := make(chan error)
	go func() {
		// watch should open up with an initial read
		list, err := rc.List(namespace, clients.ListOpts{
			Ctx:      opts.Ctx,
			Selector: opts.Selector,
		})
		if err != nil {
			errs <- err
			return
		}
		resourcesChan <- list
		for {
			select {
			case <-time.After(opts.RefreshRate):
				list, err := rc.List(namespace, clients.ListOpts{
					Ctx: opts.Ctx,
				})
				if err != nil {
					errs <- err
				}
				resourcesChan <- list
			case <-opts.Ctx.Done():
				close(resourcesChan)
				close(errs)
				return
			}
		}
	}()

	return resourcesChan, errs, nil
}

const (
	directoryTypeData     = "data"
	directoryTypeMetadata = "metadata"
)

func (rc *ResourceClient) resourceDirectory(namespace, directoryType string) string {
	return strings.Join([]string{
		rc.pathPrefix,
		directoryType,
		rc.root,
		rc.resourceType.GroupVersionKind().Group,
		rc.resourceType.GroupVersionKind().Version,
		rc.resourceType.GroupVersionKind().Kind,
		namespace,
	}, "/")
}

func (rc *ResourceClient) resourceKey(namespace, name string) string {
	return strings.Join([]string{
		rc.resourceDirectory(namespace, directoryTypeData),
		name}, "/")
}

func (rc *ResourceClient) resourceMetadataKey(namespace, name string) string {
	return strings.Join([]string{
		rc.resourceDirectory(namespace, directoryTypeMetadata),
		name}, "/")
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// package main

// import (
// 	"context"
// 	"fmt"
// 	"log"

// 	vault "github.com/hashicorp/vault/api"
// 	auth "github.com/hashicorp/vault/api/auth/userpass"
// )

// Once you've set the token for your Vault client, you will need to
// periodically renew its lease.
//
// A function like this should be run as a goroutine to avoid blocking.
//
// Production applications may also wish to be more tolerant of failures and
// retry rather than exiting.
//
// Additionally, enterprise Vault users should be aware that due to eventual
// consistency, the API may return unexpected errors when running Vault with
// performance standbys or performance replication, despite the client having
// a freshly renewed token. See https://www.vaultproject.io/docs/enterprise/consistency#vault-1-7-mitigations
// for several ways to mitigate this which are outside the scope of this code sample.
func renewToken(client *vault.Client) {
	for {
		vaultLoginResp, err := login(client)
		if err != nil {
			log.Fatalf("unable to authenticate to Vault: %v", err)
		}
		tokenErr := manageTokenLifecycle(client, vaultLoginResp)
		if tokenErr != nil {
			log.Fatalf("unable to start managing token lifecycle: %v", tokenErr)
		}
	}
}

// Starts token lifecycle management. Returns only fatal errors as errors,
// otherwise returns nil so we can attempt login again.
func manageTokenLifecycle(client *vault.Client, token *vault.Secret) error {
	renew := token.Auth.Renewable // You may notice a different top-level field called Renewable. That one is used for dynamic secrets renewal, not token renewal.
	if !renew {
		log.Printf("Token is not configured to be renewable. Re-attempting login.")
		return nil
	}

	watcher, err := client.NewLifetimeWatcher(&vault.LifetimeWatcherInput{
		Secret:    token,
		Increment: 3600, // Learn more about this optional value in https://www.vaultproject.io/docs/concepts/lease#lease-durations-and-renewal
	})
	if err != nil {
		return fmt.Errorf("unable to initialize new lifetime watcher for renewing auth token: %w", err)
	}

	go watcher.Start()
	defer watcher.Stop()

	for {
		select {
		// `DoneCh` will return if renewal fails, or if the remaining lease
		// duration is under a built-in threshold and either renewing is not
		// extending it or renewing is disabled. In any case, the caller
		// needs to attempt to log in again.
		case err := <-watcher.DoneCh():
			if err != nil {
				log.Printf("Failed to renew token: %v. Re-attempting login.", err)
				return nil
			}
			// This occurs once the token has reached max TTL.
			log.Printf("Token can no longer be renewed. Re-attempting login.")
			return nil

		// Successfully completed renewal
		case renewal := <-watcher.RenewCh():
			log.Printf("Successfully renewed: %#v", renewal)
		}
	}
}

func login(client *vault.Client) (*vault.Secret, error) {
	// WARNING: A plaintext password like this is obviously insecure.
	// See the files in the auth-methods directory for full examples of how to securely
	// log in to Vault using various auth methods. This function is just
	// demonstrating the basic idea that a *vault.Secret is returned by
	// the login call.
	userpassAuth, err := auth.NewUserpassAuth("my-user", &auth.Password{FromString: "my-password"})
	if err != nil {
		return nil, fmt.Errorf("unable to initialize userpass auth method: %w", err)
	}

	authInfo, err := client.Auth().Login(context.Background(), userpassAuth)
	if err != nil {
		return nil, fmt.Errorf("unable to login to userpass auth method: %w", err)
	}
	if authInfo == nil {
		return nil, fmt.Errorf("no auth info was returned after login")
	}

	return authInfo, nil
}
