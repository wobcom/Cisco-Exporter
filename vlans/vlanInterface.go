package vlans

// VLANInterface represents a VLAN interface on cisco devices.
type VLANInterface struct {
	Name string

	InputBytes  float64
	OutputBytes float64
}
