package permissions

import (
	"fmt"
	"strings"
)

// PermissionsArray describes a set of permission
// rules.
//
// Example:
//   +sp.guild.config.*
//   +sp.*
//   +sp.guild.*
//   -sp.guild.mod.ban
//   +sp.etc.*
//   +sp.chat.*
type PermissionArray []string

// Updates "adds" the passed newPerm to the permission array
// p by merging the permissions and returns the result as
// new permission array.
//
// This means, if p looks like following
//   +sp.guild.*
//   +sp.guild.mod.ban
// and newPerm is '-sp.guild.mod.ban', the
// returned permission array will be
//   +sp.guild.*
func (p PermissionArray) Update(newPerm string, override bool) PermissionArray {
	newPermsArray := make(PermissionArray, len(p)+1)

	i := 0
	add := true
	for _, perm := range p {
		// If the permission rule equals and overrride
		// is true, set newPerm at this point to
		// newPermsArray.
		// Otherwise, if the prefix of perm and newPerm
		// are unequal and prefix of newPerm is '-',
		// the is not being added.
		// If prefix of perm and newPerm are unequal and
		// prefix of newPerm is '+', newPerm will be added
		// to newPermsArray.
		//
		// Otherwise, perm is added to newPermArray.
		if len(perm) > 0 && perm[1:] == newPerm[1:] {
			add = false

			if override {
				newPermsArray[i] = newPerm
				i++
				continue
			}

			if perm[0] != newPerm[0] {
				if newPerm[0] == '-' {
					continue
				} else {
					newPermsArray[i] = newPerm
					i++
				}
			}
		}

		newPermsArray[i] = perm
		i++
	}

	if add {
		newPermsArray[i] = newPerm
		i++
	}

	return newPermsArray[:i]
}

// Merge updates all entries of p using Update one
// by one with all entries of newPerms. Parameter
// override is passed to the Update function.
//
// A new permissions array is returned with the
// resulting permission rule set.
func (p PermissionArray) Merge(newPerms PermissionArray, override bool) PermissionArray {
	for _, cp := range newPerms {
		p = p.Update(cp, override)
	}
	return p
}

// Check returns true if the passed domainName
// matches positively on the permission array p.
func (p PermissionArray) Check(domainName string) bool {
	lvl := -1
	allow := false

	for _, perm := range p {
		m, a := permissionCheckDNs(domainName, perm)
		if m > lvl {
			allow = a
			lvl = m
		}
	}

	return allow
}

// permissionMatchDNs tries to match the passed
// domainName on the passed perm.
//
// This also respects explicit domainNames
// prefixed with '!'.
//
// The resulting match index is returned. If the
// match index is < 0, this must be interpreted as
// no match.
func permissionMatchDNs(domainName, perm string) int {
	var needsExplicitAllow bool

	// A domainName with the prefix '!' sets
	// needsExplicitAllow to true.
	// This means, the domainName must be
	// explicitely allowed and can not be matched
	// by wildcard.
	if domainName[0] == '!' {
		needsExplicitAllow = true
		domainName = domainName[1:]
	}

	// If the domain name equals perm, return
	// 999 match index.
	if domainName == perm {
		return 999
	}

	// ...otherwise, if needsExplicitAllow is
	// true and it is not an exact match,
	// return negative match.
	if needsExplicitAllow {
		return -1
	}

	// Split domainName in areas seperated by '.'
	dnAreas := strings.Split(domainName, ".")
	assembled := ""
	for i, dnArea := range dnAreas {
		if assembled == "" {
			// If assembled is empty, set assembled to
			// current dnArea.
			assembled = dnArea
		} else {
			// Otherwise, add current dnArea to assembled.
			assembled = fmt.Sprintf("%s.%s", assembled, dnArea)
		}

		// If perm equals assembled area with trailing
		// wildcard selector ".*", return current index
		// as match index.
		if perm == fmt.Sprintf("%s.*", assembled) {
			return i
		}
	}

	// Otherwise, return negative match index.
	return -1
}

// permissionCheckDNs tries to match domainName on the
// passed perm and returns the match index and if
// it matched and perm is not prefixed with '-'.
func permissionCheckDNs(domainName, perm string) (int, bool) {
	match := permissionMatchDNs(domainName, perm[1:])
	if match < 0 {
		return match, false
	}

	return match, !strings.HasPrefix(perm, "-")
}
