package uam

import "github.com/go-ldap/ldap/v3"

type GroupInfo struct {
	Group   *ldap.Entry
	Members []*ldap.Entry
}
