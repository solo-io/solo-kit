package matchers

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-test/deep"
	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/proto"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
)

// Use this in a gomock EXPECT call e.g.
// `client.EXPECT().Update(ctx, GomockMatchProto(expected)).Return(nil)`
func GomockMatchProto(actual interface{}) gomock.Matcher {
	return &gomockProtoMatcher{
		actual: actual,
	}
}

type gomockProtoMatcher struct {
	actual interface{}
	diff   []string
}

func (p *gomockProtoMatcher) Matches(actual interface{}) bool {
	diff := deep.Equal(p.actual, actual)
	p.diff = diff
	return len(diff) == 0
}

func (p *gomockProtoMatcher) String() string {
	return fmt.Sprintf("equals proto %v", p.actual)
}

// Use this in an Expect call e.g.
// Expect(result).To(MatchProto(expected))
func MatchProto(msg proto.Message) types.GomegaMatcher {
	return &protoMatcherImpl{
		msg: msg,
	}
}

type protoMatcherImpl struct {
	msg proto.Message
}

func (p *protoMatcherImpl) Match(actual interface{}) (success bool, err error) {
	msg, ok := actual.(proto.Message)
	if !ok {
		return false, nil
	}
	return proto.Equal(msg, p.msg), nil
}

func (p *protoMatcherImpl) FailureMessage(actual interface{}) (message string) {
	msg, ok := actual.(proto.Message)
	if !ok {
		format.Message(actual, "To be identical to", p.msg.String())
	}
	return format.Message(msg.String(), "To be identical to", p.msg.String())
}

func (p *protoMatcherImpl) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "Not to be identical to", p.msg)
}

// Use this in an Expect call e.g.
// Expect(result).To(MatchProto(expected))
// especially when the `expected` type is an skv2 type with no String() function implemented.
func MatchPublicFields(actual interface{}) types.GomegaMatcher {
	return &publicFieldMatcher{actual: actual}
}

type publicFieldMatcher struct {
	actual interface{}
	diff   []string
}

func (p *publicFieldMatcher) Match(actual interface{}) (success bool, err error) {
	diff := deep.Equal(p.actual, actual)
	p.diff = diff
	return len(diff) == 0, nil
}

func (p *publicFieldMatcher) FailureMessage(actual interface{}) (message string) {
	return strings.Join(p.diff, " ")
}

func (p *publicFieldMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return "not equal"
}

func ContainProto(msg proto.Message) types.GomegaMatcher {
	return &protoContainImpl{
		msg: msg,
	}
}

type protoContainImpl struct {
	msg proto.Message
}

func (p *protoContainImpl) Match(actual interface{}) (success bool, err error) {
	protoList, ok := interfaceToProtoList(actual)
	if !ok {
		return false, nil
	}

	for _, incoming := range protoList {
		if proto.Equal(incoming, p.msg) {
			return true, nil
		}
	}

	return false, nil
}

func (p *protoContainImpl) FailureMessage(actual interface{}) (message string) {
	protoList, ok := interfaceToProtoList(actual)
	if !ok {
		format.Message(actual, "Not of the correct type, expected a list of protos")
	}
	return format.Message(protoList, "To contain to", p.msg.String())
}

func (p *protoContainImpl) NegatedFailureMessage(actual interface{}) (message string) {
	protoList, ok := interfaceToProtoList(actual)
	if !ok {
		format.Message(actual, "Not of the correct type, expected a list of protos")
	}
	return format.Message(protoList, "Not to contain ", p.msg.String())
}

func ConsistOfProtos(msgs ...proto.Message) types.GomegaMatcher {
	return &protoConsist{
		msgs: msgs,
	}
}

type protoConsist struct {
	msgs []proto.Message
}

func (p *protoConsist) Match(actual interface{}) (success bool, err error) {

	protoList, ok := interfaceToProtoList(actual)
	if !ok {
		return false, nil
	}

	for _, incoming := range protoList {
		var contains bool
		for _, present := range p.msgs {
			if proto.Equal(present, incoming) {
				contains = true
				break
			}
		}
		if !contains {
			return false, nil
		}
	}
	return true, nil
}

func (p *protoConsist) FailureMessage(actual interface{}) (message string) {
	protoList, ok := interfaceToProtoList(actual)
	if !ok {
		format.Message(actual, "Not of the correct type, expected a list of protos")
	}
	return format.Message(stringProtoList(protoList), "To consist of ", p.msgs)
}

func (p *protoConsist) NegatedFailureMessage(actual interface{}) (message string) {
	protoList, ok := interfaceToProtoList(actual)
	if !ok {
		format.Message(actual, "Not of the correct type, expected a list of protos")
	}
	return format.Message(stringProtoList(protoList), "Not to consist of ", p.msgs)
}

func interfaceToProtoList(actual interface{}) (result []proto.Message, ok bool) {
	switch reflect.TypeOf(actual).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(actual)
		for i := 0; i < s.Len(); i++ {
			msg, isProto := s.Index(i).Interface().(proto.Message)
			if !isProto {
				return
			}
			result = append(result, msg)
		}
	}

	return result, true
}

func stringProtoList(msgs []proto.Message) (result []string) {
	for _, v := range msgs {
		result = append(result, v.String())
	}
	return
}
