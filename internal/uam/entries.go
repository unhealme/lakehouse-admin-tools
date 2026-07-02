package uam

import (
	"encoding/csv"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-ldap/ldap/v3"
)

const defaultFmt = "%-18s : %s\n"

type PrintFormat int

const (
	PrintFormatDefault PrintFormat = iota + 1
	PrintFormatCSV
)

func parseGroup(values []string, base string) string {
	var groups []string
	for _, grp := range values {
		if strings.Contains(grp, base) {
			cn, _, _ := strings.Cut(grp, ",")
			_, group, _ := strings.Cut(cn, "=")
			groups = append(groups, group)
		}
	}
	slices.Sort(groups)
	return strings.Join(groups, ",")
}

func parseTime(ldapTime string) string {
	t, err := strconv.ParseInt(ldapTime, 10, 64)
	if err != nil {
		return strconv.FormatInt(t, 10)
	} else if t == 0 {
		return "0"
	}
	return time.Unix((t/10000000)-11644473600, 0).Local().String()
}

func PrintDefault(entry *ldap.Entry, groupBase string) {
	fmt.Printf(defaultFmt, "distinguishedName", entry.DN)
	for _, attr := range entry.Attributes {
		switch attr.Name {
		case "extensionAttribute13":
			fmt.Printf(defaultFmt, "directorate", attr.Values[0])
		case "extensionAttribute14":
			fmt.Printf(defaultFmt, "divisionGroup", attr.Values[0])
		case "extensionAttribute15":
			fmt.Printf(defaultFmt, "division", attr.Values[0])
		case "sAMAccountName":
			fmt.Printf(defaultFmt, "username", attr.Values[0])
		case "memberOf":
			fmt.Printf(defaultFmt, "group", parseGroup(attr.Values, groupBase))
		case "badPasswordTime", "lockoutTime", "pwdLastSet", "lastLogon":
			fmt.Printf(defaultFmt, attr.Name, parseTime(attr.Values[0]))
		default:
			fmt.Printf(defaultFmt, attr.Name, attr.Values[0])
		}
	}
}

func PrintCSV(writer *csv.Writer, entry *ldap.Entry, groupBase string) {
	// name,username,mail,department,directorate,divisionGroup,division,group,distinguishedName,badPwdCount,badPasswordTime,lockoutTime,pwdLastSet,lastLogon
	writer.Write(
		[]string{
			entry.GetAttributeValue("name"),
			entry.GetAttributeValue("sAMAccountName"),
			entry.GetAttributeValue("mail"),
			entry.GetAttributeValue("department"),
			entry.GetAttributeValue("extensionAttribute13"),
			entry.GetAttributeValue("extensionAttribute14"),
			entry.GetAttributeValue("extensionAttribute15"),
			parseGroup(entry.GetAttributeValues("memberOf"), groupBase),
			entry.DN,
			entry.GetAttributeValue("badPwdCount"),
			parseTime(entry.GetAttributeValue("badPasswordTime")),
			parseTime(entry.GetAttributeValue("lockoutTime")),
			parseTime(entry.GetAttributeValue("pwdLastSet")),
			parseTime(entry.GetAttributeValue("lastLogon")),
		})
}
