package uam

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/url"

	ldap "github.com/go-ldap/ldap/v3"
	"github.com/go-ldap/ldap/v3/gssapi"
	"github.com/jcmturner/gokrb5/v8/client"
	"github.com/jcmturner/gokrb5/v8/iana/flags"
	"github.com/unhealme/lakehouse-admin-tools/internal"
)

type UamClient struct {
	*ldap.Conn
	baseDn, groupBase, domain string
}

func (c UamClient) DescribeUser(user string, printFmt PrintFormat, writer *csv.Writer) error {
	req := ldap.NewSearchRequest(
		c.baseDn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=user)(|(sAMAccountName=%[1]s)(mail=%[1]s@%[2]s)(mail=%[1]s)))", user, c.domain),
		[]string{
			"badPasswordTime",
			"badPwdCount",
			"department",
			"dn",
			"extensionAttribute13", // directorate
			"extensionAttribute14", // division group
			"extensionAttribute15", // division
			"lastLogon",
			"lockoutTime",
			"mail",
			"memberOf",
			"name",
			"pwdLastSet",
			"sAMAccountName",
		},
		nil,
	)
	result, err := c.Search(req)
	if err != nil {
		return err
	}
	found := false
	for _, entry := range result.Entries {
		switch printFmt {
		case PrintFormatDefault:
			PrintDefault(entry, c.groupBase)
		case PrintFormatCSV:
			PrintCSV(writer, entry, c.groupBase)
		}
		found = true
	}
	if !found {
		return errors.New("user not found")
	}
	return nil
}

func (c UamClient) ListMembers(group string) error {
	req := ldap.NewSearchRequest(
		c.baseDn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=group)(sAMAccountName=%s))", group),
		[]string{"cn", "member"},
		nil,
	)
	result, err := c.Search(req)
	if err != nil {
		return err
	}
	found := false
	for _, entry := range result.Entries {
		fmt.Printf("%s : %s\n", entry.GetAttributeValue("cn"), parseGroup(entry.GetAttributeValues("member"), ""))
		found = true
	}
	if !found {
		return errors.New("group not found")
	}
	return nil
}

func NewClient(
	logger *internal.Logger, ldapUrl, user, passw string,
	baseDn, groupBase, domain, realm string,
) (*UamClient, error) {
	base, err := ldap.DialURL(ldapUrl)
	if err != nil {
		return nil, err
	}
	if err := base.Bind(user, passw); err != nil {
		logger.Debug("ldap bind failed. trying gssapi bind..", logger.Args("error", err))

		// https://github.com/go-ldap/ldap/issues/536
		gssapiClient, err := gssapi.NewClientWithPassword(
			user,
			realm,
			passw,
			internal.GetEnv("KRB5_CONFIG", "/etc/krb5.conf"),
			client.DisablePAFXFAST(true),
		)
		if err != nil {
			return nil, err
		}
		defer gssapiClient.Close()
		parsedUrl, err := url.Parse(ldapUrl)
		if err != nil {
			return nil, err
		}

		// https://github.com/go-ldap/ldap/issues/536#issuecomment-2473581901
		bindReq := &ldap.GSSAPIBindRequest{}
		bindReq.ServicePrincipalName = fmt.Sprintf("ldap/%s", parsedUrl.Hostname())
		if err := base.GSSAPIBindRequestWithAPOptions(gssapiClient, bindReq, []int{flags.APOptionMutualRequired}); err != nil {
			return nil, err
		}
	}
	return &UamClient{base, baseDn, groupBase, domain}, nil
}
