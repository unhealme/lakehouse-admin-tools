package uam

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-ldap/ldap/v3"
)

const defaultFmt = "%-18s : %s\n"

func parseTime(t int64) string {
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
			var groups []string
			for _, grp := range attr.Values {
				if strings.Contains(grp, groupBase) {
					cn, _, _ := strings.Cut(grp, ",")
					_, group, _ := strings.Cut(cn, "=")
					groups = append(groups, group)
				}
			}
			slices.Sort(groups)
			fmt.Printf(defaultFmt, "group", strings.Join(groups, ","))
		case "badPasswordTime", "lockoutTime", "pwdLastSet", "lastLogon":
			switch attr.Values[0] {
			case "0":
				fmt.Printf(defaultFmt, attr.Name, attr.Values[0])
			default:
				i, _ := strconv.ParseInt(attr.Values[0], 10, 64)
				fmt.Printf(defaultFmt, attr.Name, parseTime(i))
			}
		default:
			fmt.Printf(defaultFmt, attr.Name, attr.Values[0])
		}
	}
}
