package types

import (
	"cosmossdk.io/collections"

	ormv1alpha1 "cosmossdk.io/api/cosmos/orm/v1alpha1"
)

var (
	// ParamsKey saves the current module params.
	ParamsKey = collections.NewPrefix(0)
)

const (
	ModuleName = "profile"

	StoreKey = ModuleName

	QuerierRoute = ModuleName

	AuthorityKeyPrefix              = "Authority/admin/"
	AuthorityEditableAdminKeyPrefix = "Authority/editable/admin/"

	ProfileKeyPrefix           = "Profile/value/"
	ProfileUserHandleKeyPrefix = "Profile/userHandle/"
	ProfileUserSearchKeyPrefix = "Profile/userSearch/"

	ProfileFollowingPrefix  = "Profile/following/"
	ProfileFollowersPrefix  = "Profile/followers/"
	ProfileFollowTimePrefix = "Profile/follow/time/"

	ActivitiesReceivedPrefix      = "Activities/received/"
	ActivitiesReceivedCountPrefix = "Activities/received/count/"

	ActivitiesReceivedCount = 100
	AdminActionAppoint      = "appoint"
	AdminActionRemove       = "remove"
)

var ORMModuleSchema = ormv1alpha1.ModuleSchemaDescriptor{
	SchemaFile: []*ormv1alpha1.ModuleSchemaDescriptor_FileEntry{
		{Id: 1, ProtoFileName: "profile/v1/state.proto"},
	},
	Prefix: []byte{0},
}
