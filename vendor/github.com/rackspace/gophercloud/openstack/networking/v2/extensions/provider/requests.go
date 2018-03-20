package provider

import (
	"github.com/rackspace/gophercloud"
)

// CreateOptsBuilder is the interface options structs have to satisfy in order
// to be used in the main Create operation in this package. Since many
// extensions decorate or modify the common logic, it is useful for them to
// satisfy a basic interface in order for them to be used.
type CreateNetOptsBuilder interface {
	ToNetworkCreateMap() (map[string]interface{}, error)
}

type CreateNetOpts struct {
	AdminStateUp    *bool
	Name            string
	Shared          *bool
	TenantID        string
	NetworkType     string
	PhysicalNetwork string
	SegmentationID  string
}

// ToNetworkCreateMap casts a CreateOpts struct to a map.
func (opts CreateNetOpts) ToNetworkCreateMap() (map[string]interface{}, error) {
	n := make(map[string]interface{})

	if opts.AdminStateUp != nil {
		n["admin_state_up"] = &opts.AdminStateUp
	}
	if opts.Name != "" {
		n["name"] = opts.Name
	}
	if opts.Shared != nil {
		n["shared"] = &opts.Shared
	}
	if opts.TenantID != "" {
		n["tenant_id"] = opts.TenantID
	}
	if opts.NetworkType != "" {
		n["provider:network_type"] = opts.NetworkType
	}
	if opts.PhysicalNetwork != "" {
		n["provider:physical_network"] = opts.PhysicalNetwork
	}
	if opts.SegmentationID != "" {
		n["provider:segmentation_id"] = opts.SegmentationID
	}

	return map[string]interface{}{"network": n}, nil
}

// Create accepts a CreateOpts struct and creates a new network using the values
// provided. This operation does not actually require a request body, i.e. the
// CreateOpts struct argument can be empty.
//
// The tenant ID that is contained in the URI is the tenant that creates the
// network. An admin user, however, has the option of specifying another tenant
// ID in the CreateOpts struct.
func CreateNet(c *gophercloud.ServiceClient, opts CreateNetOptsBuilder) CreateResult {
	var res CreateResult

	reqBody, err := opts.ToNetworkCreateMap()
	if err != nil {
		res.Err = err
		return res
	}

	_, res.Err = c.Post(createURL(c), reqBody, &res.Body, nil)
	return res
}

// CreateOptsBuilder is the interface options structs have to satisfy in order
// to be used in the main Create operation in this package. Since many
// extensions decorate or modify the common logic, it is useful for them to
// satisfy a basic interface in order for them to be used.
type CreatePortOptsBuilder interface {
	ToPortCreateMap() (map[string]interface{}, error)
}

// CreateOpts represents the attributes used when creating a new port.
type CreatePortOpts struct {
	NetworkID           string
	Name                string
	AdminStateUp        *bool
	MACAddress          string
	FixedIPs            interface{}
	DeviceID            string
	DeviceOwner         string
	TenantID            string
	SecurityGroups      []string
	AllowedAddressPairs []AddressPair
	HostId              string
	Profile             string
	VnicType            string
}

// ToPortCreateMap casts a CreateOpts struct to a map.
func (opts CreatePortOpts) ToPortCreateMap() (map[string]interface{}, error) {
	p := make(map[string]interface{})

	if opts.NetworkID == "" {
		return nil, errNetworkIDRequired
	}
	p["network_id"] = opts.NetworkID

	if opts.DeviceID != "" {
		p["device_id"] = opts.DeviceID
	}
	if opts.DeviceOwner != "" {
		p["device_owner"] = opts.DeviceOwner
	}
	if opts.FixedIPs != nil {
		p["fixed_ips"] = opts.FixedIPs
	}
	if opts.SecurityGroups != nil {
		p["security_groups"] = opts.SecurityGroups
	}
	if opts.TenantID != "" {
		p["tenant_id"] = opts.TenantID
	}
	if opts.AdminStateUp != nil {
		p["admin_state_up"] = &opts.AdminStateUp
	}
	if opts.Name != "" {
		p["name"] = opts.Name
	}
	if opts.MACAddress != "" {
		p["mac_address"] = opts.MACAddress
	}
	if opts.AllowedAddressPairs != nil {
		p["allowed_address_pairs"] = opts.AllowedAddressPairs
	}
	if opts.HostId != "" {
		p["binding:host_id"] = opts.HostId
	}
	if opts.Profile != "" {
		p["binding:profile"] = opts.Profile
	}
	if opts.VnicType != "" {
		p["binding:vnic_type"] = opts.VnicType
	}

	return map[string]interface{}{"port": p}, nil
}

// Create accepts a CreateOpts struct and creates a new network using the values
// provided. You must remember to provide a NetworkID value.
func CreatePort(c *gophercloud.ServiceClient, opts CreatePortOptsBuilder) CreateResult {
	var res CreateResult

	reqBody, err := opts.ToPortCreateMap()
	if err != nil {
		res.Err = err
		return res
	}

	_, res.Err = c.Post(createURL(c), reqBody, &res.Body, nil)
	return res
}
