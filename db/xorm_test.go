/*
 * @Author:   admin
 * @IDE:      GoLand
 * @Date:     2025/8/1 14:12
 * @FilePath: qingliu/db/xorm_test.go
 */

package db

import (
	"context"
	"testing"

	"xorm.io/xorm"
)

func TestName(t *testing.T) {
	err := Transaction(context.Background(), "default", func(session *xorm.Session) error {
		// 执行事务操作
		if _, err := session.Insert(&user); err != nil {
			return err
		}
		if _, err := session.Update(&account); err != nil {
			return err
		}
		return nil
	})
}
