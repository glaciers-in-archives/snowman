package version

import "strconv"

type Version struct {
	Major  int
	Minor  int
	Patch  int
	Suffix string
}

func (v Version) String() string {
	versionNumber := strconv.Itoa(v.Major) + "." + strconv.Itoa(v.Minor) + "." + strconv.Itoa(v.Patch)
	if !(v.Suffix == "") {
		versionNumber += "-" + v.Suffix
	}
	return versionNumber
}
