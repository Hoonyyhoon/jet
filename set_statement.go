package jet

import (
	"errors"
)

// UNION effectively appends the result of sub-queries(select statements) into single query.
// It eliminates duplicate rows from its result.
func UNION(lhs, rhs SelectStatement, selects ...SelectStatement) SelectStatement {
	return newSetStatementImpl(union, false, toSelectList(lhs, rhs, selects...))
}

// UNION_ALL effectively appends the result of sub-queries(select statements) into single query.
// It does not eliminates duplicate rows from its result.
func UNION_ALL(lhs, rhs SelectStatement, selects ...SelectStatement) SelectStatement {
	return newSetStatementImpl(union, true, toSelectList(lhs, rhs, selects...))
}

// INTERSECT returns all rows that are in query results.
// It eliminates duplicate rows from its result.
func INTERSECT(lhs, rhs SelectStatement, selects ...SelectStatement) SelectStatement {
	return newSetStatementImpl(intersect, false, toSelectList(lhs, rhs, selects...))
}

// INTERSECT_ALL returns all rows that are in query results.
// It does not eliminates duplicate rows from its result.
func INTERSECT_ALL(lhs, rhs SelectStatement, selects ...SelectStatement) SelectStatement {
	return newSetStatementImpl(intersect, true, toSelectList(lhs, rhs, selects...))
}

// EXCEPT returns all rows that are in the result of query lhs but not in the result of query rhs.
// It eliminates duplicate rows from its result.
func EXCEPT(lhs, rhs SelectStatement) SelectStatement {
	return newSetStatementImpl(except, false, toSelectList(lhs, rhs))
}

// EXCEPT_ALL returns all rows that are in the result of query lhs but not in the result of query rhs.
// It does not eliminates duplicate rows from its result.
func EXCEPT_ALL(lhs, rhs SelectStatement) SelectStatement {
	return newSetStatementImpl(except, true, toSelectList(lhs, rhs))
}

func toSelectList(lhs, rhs SelectStatement, selects ...SelectStatement) []SelectStatement {
	return append([]SelectStatement{lhs, rhs}, selects...)
}

const (
	union     = "UNION"
	intersect = "INTERSECT"
	except    = "EXCEPT"
)

// Similar to selectStatementImpl, but less complete
type setStatementImpl struct {
	selectStatementImpl

	operator string
	all      bool
	selects  []SelectStatement
}

func newSetStatementImpl(operator string, all bool, selects []SelectStatement) SelectStatement {
	setStatement := &setStatementImpl{
		operator: operator,
		all:      all,
		selects:  selects,
	}

	setStatement.selectStatementImpl.expressionInterfaceImpl.parent = setStatement
	setStatement.selectStatementImpl.parent = setStatement
	setStatement.limit = -1
	setStatement.offset = -1

	return setStatement
}

func (s *setStatementImpl) projections() []projection {
	if len(s.selects) > 0 {
		return s.selects[0].projections()
	}
	return []projection{}
}

func (s *setStatementImpl) serialize(statement statementType, out *sqlBuilder, options ...serializeOption) error {
	if s == nil {
		return errors.New("jet: Set expression is nil. ")
	}

	wrap := s.orderBy != nil || s.limit >= 0 || s.offset >= 0

	if wrap {
		out.writeString("(")
		out.increaseIdent()
	}

	err := s.serializeImpl(out)

	if err != nil {
		return err
	}

	if wrap {
		out.decreaseIdent()
		out.newLine()
		out.writeString(")")
	}

	return nil
}

func (s *setStatementImpl) serializeImpl(out *sqlBuilder) error {
	if s == nil {
		return errors.New("jet: Set expression is nil. ")
	}

	if len(s.selects) < 2 {
		return errors.New("jet: UNION Statement must have at least two SELECT statements")
	}

	out.newLine()
	out.writeString("(")
	out.increaseIdent()

	for i, selectStmt := range s.selects {
		out.newLine()
		if i > 0 {
			out.writeString(s.operator)

			if s.all {
				out.writeString("ALL")
			}
			out.newLine()
		}

		if selectStmt == nil {
			return errors.New("jet: select statement is nil")
		}

		err := selectStmt.serialize(setStatement, out)

		if err != nil {
			return err
		}
	}

	out.decreaseIdent()
	out.newLine()
	out.writeString(")")

	if s.orderBy != nil {
		err := out.writeOrderBy(setStatement, s.orderBy)
		if err != nil {
			return err
		}
	}

	if s.limit >= 0 {
		out.newLine()
		out.writeString("LIMIT")
		out.insertParametrizedArgument(s.limit)
	}

	if s.offset >= 0 {
		out.newLine()
		out.writeString("OFFSET")
		out.insertParametrizedArgument(s.offset)
	}

	return nil
}

func (s *setStatementImpl) Sql() (query string, args []interface{}, err error) {
	queryData := &sqlBuilder{}

	err = s.serializeImpl(queryData)

	if err != nil {
		return
	}

	query, args = queryData.finalize()
	return
}
