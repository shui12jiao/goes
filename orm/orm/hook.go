package orm

const (
	BeforeQuery = iota
	AfterQuery
	BeforeInsert
	AfterInsert
	BeforeUpdate
	AfterUpdate
	BeforeDelete
	AfterDelete
)

type IBeforeQuery interface {
	BeforeQuery(*Session) error
}

type IAfterQuery interface {
	AfterQuery(*Session) error
}

type IBeforeInsert interface {
	BeforeInsert(*Session) error
}

type IAfterInsert interface {
	AfterInsert(*Session) error
}

type IBeforeUpdate interface {
	BeforeUpdate(*Session) error
}

type IAfterUpdate interface {
	AfterUpdate(*Session) error
}

type IBeforeDelete interface {
	BeforeDelete(*Session) error
}

type IAfterDelete interface {
	AfterDelete(*Session) error
}
