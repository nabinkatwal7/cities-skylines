package ui

import (
	"fmt"
	"strings"

	"github.com/katwate/js-skylines/internal/transport"
	"github.com/katwate/js-skylines/internal/zoning"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type AssetFlags uint8

const (
	AssetFavorite AssetFlags = 1 << iota
	AssetDLC
	AssetMod
	AssetRecent
)

// BuildAsset is one placeable entry in a build category (24.4).
type BuildAsset struct {
	ID           string
	Name         string
	Category     ToolbarCategory
	Cost         float32
	Maintenance  float32
	Size         string
	Requirements string
	UnlockPop    int
	Flags        AssetFlags
	Preview      string // ponytail: text preview until 3D thumbnails exist
}

// BuildMenus lists placeable assets per build category (24.4).
type BuildMenus struct {
	open       bool
	category   ToolbarCategory
	assets     []BuildAsset
	filtered   []int
	selected   int
	search     string
	showFav    bool
	showRecent bool
	recent     []string
	favorites  map[string]bool
	unlocks    *UnlockRegistry
}

func NewBuildMenus(unlocks *UnlockRegistry) *BuildMenus {
	return &BuildMenus{
		favorites: make(map[string]bool),
		unlocks:   unlocks,
	}
}

func (b *BuildMenus) catalog() []BuildAsset {
	if len(b.assets) > 0 {
		return b.assets
	}
	b.assets = defaultAssetCatalog()
	return b.assets
}

func defaultAssetCatalog() []BuildAsset {
	var out []BuildAsset
	for i, name := range roadOptions().Options {
		out = append(out, BuildAsset{
			ID: "road_" + name, Name: name, Category: CatRoads,
			Cost: 100 + float32(i)*50, Maintenance: 2, Size: "1 cell",
			Requirements: "Road access", UnlockPop: 0, Preview: name + " road segment",
		})
	}
	zones := zoneOptions().Options
	for i, name := range zones {
		out = append(out, BuildAsset{
			ID: "zone_" + name, Name: name, Category: CatZoning,
			Cost: 0, Maintenance: 0, Size: "4x4", Requirements: "Road frontage",
			UnlockPop: 0, Preview: name + " zoning",
		})
		_ = zoning.ZoneType(i + 1)
	}
	transportOpts := transportOptions().Options
	for i, name := range transportOpts {
		out = append(out, BuildAsset{
			ID: "trans_" + name, Name: name, Category: CatPublicTransport,
			Cost: 500, Maintenance: 10, Size: "1 tile", Requirements: "Road nearby",
			UnlockPop: 0, Preview: name + " stop",
		})
		_ = transport.TransportType(i)
	}
	serviceAssets := []struct {
		cat  ToolbarCategory
		name string
		cost float32
		pop  int
	}{
		{CatElectricity, "Coal Plant", 12000, 100},
		{CatElectricity, "Wind Turbine", 6000, 200},
		{CatWater, "Water Tower", 8000, 150},
		{CatWater, "Sewage Outlet", 5000, 180},
		{CatGarbage, "Landfill", 4000, 200},
		{CatHealthcare, "Clinic", 10000, 300},
		{CatHealthcare, "Hospital", 25000, 800},
		{CatFireRescue, "Fire Station", 8000, 400},
		{CatPolice, "Police Station", 9000, 500},
		{CatEducation, "Elementary School", 12000, 600},
		{CatEducation, "High School", 20000, 1000},
		{CatParks, "Small Park", 3000, 350},
		{CatParks, "Plaza", 5000, 500},
		{CatLandscaping, "Trees", 200, 250},
		{CatDistricts, "District Tool", 0, 50},
	}
	for _, s := range serviceAssets {
		out = append(out, BuildAsset{
			ID: strings.ToLower(strings.ReplaceAll(s.name, " ", "_")),
			Name: s.name, Category: s.cat, Cost: s.cost, Maintenance: s.cost * 0.02,
			Size: "4x4", Requirements: "Road access", UnlockPop: s.pop,
			Preview: s.name,
		})
	}
	out = append(out, BuildAsset{
		ID: "policy_tax", Name: "Tax Policy", Category: CatPolicies,
		Cost: 0, Maintenance: 0, Size: "—", Requirements: "Town hall",
		UnlockPop: 1000, Preview: "Adjust tax rates",
	})
	out = append(out, BuildAsset{
		ID: "econ_budget", Name: "Budget Panel", Category: CatEconomy,
		Cost: 0, Maintenance: 0, Size: "—", Requirements: "Treasury",
		UnlockPop: 800, Preview: "City budget overview",
	})
	return out
}

func (b *BuildMenus) OpenCategory(cat ToolbarCategory) {
	b.category = cat
	b.open = UsesBuildMenu(cat) || cat == CatEconomy || cat == CatPolicies
	b.selected = 0
	b.refilter()
}

func (b *BuildMenus) Visible() bool { return b.open }

func (b *BuildMenus) Close() { b.open = false }

func (b *BuildMenus) Y() int32 {
	y := ToolbarY - OptionsBarH
	if b.open {
		y -= BuildMenuH
	}
	return int32(y)
}

func (b *BuildMenus) refilter() {
	b.filtered = b.filtered[:0]
	q := strings.ToLower(strings.TrimSpace(b.search))
	for i, a := range b.catalog() {
		if a.Category != b.category {
			continue
		}
		if b.unlocks != nil && b.unlocks.Population < a.UnlockPop {
			continue
		}
		if b.showFav && !b.favorites[a.ID] {
			continue
		}
		if b.showRecent {
			if !containsStr(b.recent, a.ID) {
				continue
			}
		}
		if q != "" && !strings.Contains(strings.ToLower(a.Name), q) {
			continue
		}
		b.filtered = append(b.filtered, i)
	}
	if b.selected >= len(b.filtered) {
		b.selected = 0
	}
}

func containsStr(ss []string, id string) bool {
	for _, s := range ss {
		if s == id {
			return true
		}
	}
	return false
}

func (b *BuildMenus) SelectAsset(idx int, ts *ToolSystem) {
	if idx < 0 || idx >= len(b.filtered) {
		return
	}
	a := b.catalog()[b.filtered[idx]]
	b.selected = idx
	b.touchRecent(a.ID)
	ts.ApplyAsset(a)
}

func (b *BuildMenus) touchRecent(id string) {
	for i, r := range b.recent {
		if r == id {
			b.recent = append(b.recent[:i], b.recent[i+1:]...)
			break
		}
	}
	b.recent = append([]string{id}, b.recent...)
	if len(b.recent) > 8 {
		b.recent = b.recent[:8]
	}
}

func (b *BuildMenus) ToggleFavorite(id string) {
	if b.favorites[id] {
		delete(b.favorites, id)
	} else {
		b.favorites[id] = true
	}
	b.refilter()
}

func (b *BuildMenus) HandleClick(mx, my int32, ts *ToolSystem) bool {
	if !b.open {
		return false
	}
	y := b.Y()
	if my < y || my >= y+BuildMenuH {
		return false
	}

	// Search box
	if my >= y+4 && my < y+24 && mx >= 8 && mx < 220 {
		return true
	}

	// Filter toggles
	if my >= y+4 && my < y+24 {
		if mx >= 230 && mx < 290 {
			b.showFav = !b.showFav
			b.refilter()
			return true
		}
		if mx >= 296 && mx < 360 {
			b.showRecent = !b.showRecent
			b.refilter()
			return true
		}
	}

	// Asset list
	listX := int32(8)
	listY := y + 28
	cellW := int32(88)
	cellH := int32(44)
	cols := 6
	for fi, ai := range b.filtered {
		col := fi % cols
		row := fi / cols
		bx := listX + int32(col)*cellW
		by := listY + int32(row)*cellH
		if mx >= bx && mx < bx+cellW-4 && my >= by && my < by+cellH-4 {
			b.SelectAsset(fi, ts)
			return true
		}
		_ = ai
	}
	return true
}

func (b *BuildMenus) HandleInput() {
	if !b.open {
		return
	}
	ch := rl.GetCharPressed()
	for ch != 0 {
		if ch == 8 && len(b.search) > 0 { // backspace
			b.search = b.search[:len(b.search)-1]
			b.refilter()
		} else if ch >= 32 && ch < 127 {
			b.search += string(rune(ch))
			b.refilter()
		}
		ch = rl.GetCharPressed()
	}
}

func (b *BuildMenus) Draw(ts *ToolSystem) {
	if !b.open {
		return
	}
	y := b.Y()
	rl.DrawRectangle(0, y, ScreenW, BuildMenuH, rl.NewColor(0, 0, 0, 175))

	// Search + filters
	rl.DrawRectangle(8, y+4, 210, 20, rl.NewColor(30, 30, 30, 220))
	searchLabel := b.search
	if searchLabel == "" {
		searchLabel = "Search assets..."
	}
	DrawUIText(searchLabel, 12, y+7, 13, rl.Gray)
	favCol := rl.NewColor(50, 50, 50, 200)
	if b.showFav {
		favCol = rl.NewColor(80, 70, 40, 220)
	}
	uiBtn(230, y+4, 58, 20, "Favs", favCol, rl.White, b.showFav)
	recCol := rl.NewColor(50, 50, 50, 200)
	if b.showRecent {
		recCol = rl.NewColor(40, 70, 80, 220)
	}
	uiBtn(296, y+4, 62, 20, "Recent", recCol, rl.White, b.showRecent)

	// Asset grid
	listY := y + 28
	cellW := int32(88)
	cellH := int32(44)
	cols := 6
	for fi, ai := range b.filtered {
		a := b.catalog()[ai]
		col := fi % cols
		row := fi / cols
		bx := int32(8) + int32(col)*cellW
		by := listY + int32(row)*cellH
		sel := fi == b.selected
		colFill := rl.NewColor(45, 45, 50, 220)
		if b.favorites[a.ID] {
			colFill = rl.NewColor(55, 50, 35, 220)
		}
		uiBtn(bx, by, cellW-4, cellH-4, a.Name, colFill, rl.White, sel)
		if a.Flags&AssetDLC != 0 {
			DrawUIText("DLC", bx+2, by+2, 10, rl.NewColor(255, 200, 80, 220))
		}
	}

	// Preview panel
	px := int32(ScreenW - 280)
	rl.DrawRectangle(px, y+4, 272, BuildMenuH-8, rl.NewColor(25, 25, 30, 220))
	if b.selected < len(b.filtered) {
		a := b.catalog()[b.filtered[b.selected]]
		DrawUIText(a.Name, px+8, y+10, 16, rl.White)
		DrawUIText(a.Preview, px+8, y+30, 12, rl.Gray)
		DrawUIText(fmtCost(a.Cost), px+8, y+48, 13, rl.NewColor(100, 220, 100, 220))
		DrawUIText(fmtMaint(a.Maintenance), px+8, y+64, 12, rl.Gray)
		DrawUIText("Size: "+a.Size, px+8, y+80, 12, rl.Gray)
		DrawUIText("Req: "+a.Requirements, px+8, y+96, 12, rl.Gray)
		if a.UnlockPop > 0 {
			DrawUIText(fmtUnlock(a.UnlockPop), px+8, y+112, 12, rl.NewColor(200, 180, 100, 200))
		}
	}
}

func fmtCost(c float32) string {
	if c == 0 {
		return "Cost: free"
	}
	return "Cost: $" + trimFloat(c)
}

func fmtMaint(m float32) string {
	if m == 0 {
		return "Maint: —"
	}
	return "Maint: $" + trimFloat(m) + "/wk"
}

func fmtUnlock(pop int) string {
	return "Unlock: pop " + itoa(pop)
}

func trimFloat(v float32) string {
	return strings.TrimSuffix(fmt.Sprintf("%.0f", v), ".0")
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [12]byte
	i := len(buf)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
