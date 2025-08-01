package db

import (
	"context"
	`fmt`
	"testing"

	"xorm.io/xorm"
)

func TestName(t *testing.T) {
	err := Transaction(context.Background(), "default", func(session *xorm.Session) error {
		// 执行事务操作
		if _, err := session.Insert(map[string]string{}); err != nil {
			return err
		}
		if _, err := session.Update(map[string]string{}); err != nil {
			return err
		}
		return nil
	})
	fmt.Println(err)
}
