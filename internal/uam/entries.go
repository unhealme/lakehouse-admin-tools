package uam

import (
	"bufio"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"

	ldap "github.com/go-ldap/ldap/v3"
)

const defaultFmt = "%-18s : %s\n"

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

func defaultPrinter(w io.Writer, format string, a ...any) {
	if _, err := fmt.Fprintf(w, format, a...); err != nil {
		panic(err)
	}
}

func PrintDefault(entry *ldap.Entry, groupBase string, writer *bufio.Writer) {
	defaultPrinter(writer, defaultFmt, "distinguishedName", entry.DN)
	for _, attr := range entry.Attributes {
		switch attr.Name {
		case "extensionAttribute13":
			defaultPrinter(writer, defaultFmt, "directorate", attr.Values[0])
		case "extensionAttribute14":
			defaultPrinter(writer, defaultFmt, "divisionGroup", attr.Values[0])
		case "extensionAttribute15":
			defaultPrinter(writer, defaultFmt, "division", attr.Values[0])
		case "sAMAccountName":
			defaultPrinter(writer, defaultFmt, "username", attr.Values[0])
		case "memberOf":
			defaultPrinter(writer, defaultFmt, "group", parseGroup(attr.Values, groupBase))
		case "badPasswordTime", "lockoutTime", "pwdLastSet", "lastLogon":
			defaultPrinter(writer, defaultFmt, attr.Name, parseTime(attr.Values[0]))
		default:
			defaultPrinter(writer, defaultFmt, attr.Name, attr.Values[0])
		}
	}
}

type EntryCsvSer struct {
	Entry     *ldap.Entry
	GroupBase string
}

func (e EntryCsvSer) SerCsv() []string {
	// name,username,mail,department,directorate,divisionGroup,division,group,distinguishedName,badPwdCount,badPasswordTime,lockoutTime,pwdLastSet,lastLogon
	return []string{
		e.Entry.GetAttributeValue("name"),
		e.Entry.GetAttributeValue("sAMAccountName"),
		e.Entry.GetAttributeValue("mail"),
		e.Entry.GetAttributeValue("department"),
		e.Entry.GetAttributeValue("extensionAttribute13"),
		e.Entry.GetAttributeValue("extensionAttribute14"),
		e.Entry.GetAttributeValue("extensionAttribute15"),
		parseGroup(e.Entry.GetAttributeValues("memberOf"), e.GroupBase),
		e.Entry.DN,
		e.Entry.GetAttributeValue("badPwdCount"),
		parseTime(e.Entry.GetAttributeValue("badPasswordTime")),
		parseTime(e.Entry.GetAttributeValue("lockoutTime")),
		parseTime(e.Entry.GetAttributeValue("pwdLastSet")),
		parseTime(e.Entry.GetAttributeValue("lastLogon")),
	}
}
