package enum

type PermissionType uint
type PermissionCategory uint

const (
	PermissionAll PermissionType = iota + 1
)

const (
	CategoryGeneral PermissionCategory = iota + 1
)

var permissionCategories = map[PermissionType]PermissionCategory{
	PermissionAll: CategoryGeneral,
}

var permissionNames = map[PermissionType]string{
	PermissionAll: "دسترسی کامل",
}

var CategoryNames = map[PermissionCategory]string{
	CategoryGeneral: "عمومی",
}

func (p PermissionType) String() string {
	if name, ok := permissionNames[p]; ok {
		return name
	}
	return "نامشخص"
}

func (c PermissionCategory) String() string {
	if name, ok := CategoryNames[c]; ok {
		return name
	}
	return "نامشخص"
}

func (p PermissionType) Category() PermissionCategory {
	if category, ok := permissionCategories[p]; ok {
		return category
	}
	return CategoryGeneral
}

func GetAllPermissionTypes() []PermissionType {
	return []PermissionType{
		PermissionAll,
	}
}

func GetAllPermissionCategories() []PermissionCategory {
	return []PermissionCategory{
		CategoryGeneral,
	}
}
