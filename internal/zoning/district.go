package zoning

type Policy uint32

const (
	PolicyHighRiseBan Policy = 1 << iota
	PolicyHeavyTrafficBan
	PolicySelfSufficientHousing
	PolicyITCluster
	PolicyOrganicProduce
)

type District struct {
	ID       uint8
	Policies Policy
}

func (zm *ZoneManager) initDistricts() {
	zm.districts = []District{{ID: 0}}
	zm.cellDistrict = make([][]uint8, zm.height)
	for z := range zm.cellDistrict {
		zm.cellDistrict[z] = make([]uint8, zm.width)
	}
}

func (zm *ZoneManager) DistrictAt(x, z int) uint8 {
	if zm.cellDistrict == nil || x < 0 || x >= zm.width || z < 0 || z >= zm.height {
		return 0
	}
	return zm.cellDistrict[z][x]
}

func (zm *ZoneManager) PoliciesAt(x, z int) Policy {
	id := zm.DistrictAt(x, z)
	if int(id) >= len(zm.districts) {
		return 0
	}
	return zm.districts[id].Policies
}

func (zm *ZoneManager) SetDistrictPolicy(district uint8, p Policy, enabled bool) {
	for int(district) >= len(zm.districts) {
		zm.districts = append(zm.districts, District{ID: uint8(len(zm.districts))})
	}
	if enabled {
		zm.districts[district].Policies |= p
	} else {
		zm.districts[district].Policies &^= p
	}
}

func (zm *ZoneManager) DistrictPolicies() []uint32 {
	out := make([]uint32, len(zm.districts))
	for i, d := range zm.districts {
		out[i] = uint32(d.Policies)
	}
	return out
}

func (zm *ZoneManager) ImportDistrictPolicies(policies []uint32) {
	zm.districts = zm.districts[:0]
	for i, p := range policies {
		zm.districts = append(zm.districts, District{ID: uint8(i), Policies: Policy(p)})
	}
	if len(zm.districts) == 0 {
		zm.initDistricts()
	}
}

func (zm *ZoneManager) ExportCellDistrict() []uint8 {
	out := make([]uint8, zm.width*zm.height)
	for z := 0; z < zm.height; z++ {
		for x := 0; x < zm.width; x++ {
			out[z*zm.width+x] = zm.DistrictAt(x, z)
		}
	}
	return out
}

func (zm *ZoneManager) ImportCellDistrict(flat []uint8) {
	if len(flat) != zm.width*zm.height {
		return
	}
	for z := 0; z < zm.height; z++ {
		for x := 0; x < zm.width; x++ {
			zm.cellDistrict[z][x] = flat[z*zm.width+x]
		}
	}
}
