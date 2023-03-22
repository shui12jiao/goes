package rpc

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

type Rats struct{ a1, a2 int }

type Cat struct {
	Name string
	Age  int
}

func (*Cat) Eat(args Rats, reply *string) error {
	*reply = strconv.Itoa(args.a1) + strconv.Itoa(args.a2)
	return nil
}
func (*Cat) sleep(Rats, *string) error {
	return nil
}

type dog struct {
	Name string
	Age  int
}

func (*dog) Eat(Rats, *string) error {
	return nil
}
func (*dog) sleep(Rats, *string) error {
	return nil
}

func TestNewService(t *testing.T) {
	s := newService(&Cat{})
	require.Equal(t, "Cat", s.name)
	require.Equal(t, 1, len(s.method))
	require.Equal(t, "Eat", s.method["Eat"].method.Name)

	//panic
	// s = newService(&dog{})
	// require.Equal(t, "dog", s.name)
}

func TestCall(t *testing.T) {
	s := newService(&Cat{})
	m := s.method["Eat"]
	argv := m.newArgv()
	replyv := m.newReply()
	argv.Set(reflect.ValueOf(Rats{1, 2}))
	err := s.call(m, argv, replyv)
	require.Nil(t, err)
	require.Equal(t, "12", replyv.Elem().String())
}
