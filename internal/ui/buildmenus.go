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
	scrollRow  int
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

func (b *BuildMenus) Y(hasOptionsBar bool) int32 {
	y := ToolbarY
	if hasOptionsBar {
		y -= OptionsBarH
	}
	if b.open {
		y -= BuildMenuH
	}
	return y
}

func (b *BuildMenus) chromeY(ts *ToolSystem) int32 {
	if ts == nil {
		return b.Y(false)
	}
	return b.Y(ts.HasOptionsBar())
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
	y := b.chromeY(ts)
	if my < y || my >= y+BuildMenuH {
		return false
	}

	// Search box
	if my >= y+4 && my < y+24 && mx >= 8 && mx < 220 {
		return true
	}

	// Filter toggles
	if my >= y+8 && my < y+34 {
		if mx >= 260 && mx < 324 {
			b.showFav = !b.showFav
			b.refilter()
			return true
		}
		if mx >= 332 && mx < 404 {
			b.showRecent = !b.showRecent
			b.refilter()
			return true
		}
	}

	// Asset list (matches virtualized draw)
	listX := int32(10)
	listY := y + 40
	cellW := int32(96)
	cellH := int32(56)
	cols := 6
	start, end := b.visibleAssetRange()
	for fi := start; fi < end; fi++ {
		col := fi % cols
		row := fi/cols - b.scrollRow
		bx := listX + int32(col)*cellW
		by := listY + int32(row)*cellH
		if mx >= bx && mx < bx+cellW-6 && my >= by && my < by+cellH-6 {
			b.SelectAsset(fi, ts)
			return true
		}
	}
	return false
}

func (b *BuildMenus) HandleInput() {
	if !b.open {
		return
	}
	wheel := rl.GetMouseWheelMove()
	if wheel != 0 {
		b.scrollRow -= int(wheel)
		if b.scrollRow < 0 {
			b.scrollRow = 0
		}
		maxRow := b.maxScrollRow()
		if b.scrollRow > maxRow {
			b.scrollRow = maxRow
		}
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

func (b *BuildMenus) maxScrollRow() int {
	cols := 6
	rows := (len(b.filtered) + cols - 1) / cols
	visibleRows := int((BuildMenuH - 40) / 56)
	if rows <= visibleRows {
		return 0
	}
	return rows - visibleRows
}

func (b *BuildMenus) visibleAssetRange() (start, end int) {
	cols := 6
	visibleRows := int((BuildMenuH - 40) / 56)
	if visibleRows < 1 {
		visibleRows = 1
	}
	start = b.scrollRow * cols
	end = start + visibleRows*cols
	if end > len(b.filtered) {
		end = len(b.filtered)
	}
	return start, end
}

func (b *BuildMenus) Draw(ts *ToolSystem) {
	if !b.open {
		return
	}
	y := b.chromeY(ts)
	drawBarBottom(y, BuildMenuH)

	csInputField(10, y+8, 240, 26)
	searchLabel := b.search
	if searchLabel == "" {
		searchLabel = "Search assets..."
	}
	drawLabel(searchLabel, 16, y+12, FontMd, csTextDim)
	csOptionBtn(260, y+8, 64, 26, "Favs", csBtnIdle, b.showFav)
	csOptionBtn(332, y+8, 72, 26, "Recent", csBtnIdle, b.showRecent)

	listY := y + 40
	cellW := int32(96)
	cellH := int32(56)
	cols := 6
	start, end := b.visibleAssetRange()
	for fi := start; fi < end; fi++ {
		ai := b.filtered[fi]
		a := b.catalog()[ai]
		col := fi % cols
		row := fi/cols - b.scrollRow
		bx := int32(10) + int32(col)*cellW
		by := listY + int32(row)*cellH
		csAssetBtn(bx, by, cellW-6, cellH-6, a.Name, b.favorites[a.ID], fi == b.selected)
		if a.Flags&AssetDLC != 0 {
			drawLabel("DLC", bx+4, by+4, FontXs, rl.NewColor(255, 200, 80, 255))
		}
	}

	previewW := int32(280)
	px := ScreenW - previewW - 8
	if px < 520 {
		px = 520
	}
	drawPanel(px, y+6, previewW, BuildMenuH-12)
	if b.selected < len(b.filtered) {
		a := b.catalog()[b.filtered[b.selected]]
		drawLabel(a.Name, px+10, y+12, FontMd, csText)
		drawLabel(a.Preview, px+10, y+30, FontXs, csTextDim)
		drawLabel(fmtCost(a.Cost), px+10, y+46, FontSm, csMoney)
		drawLabel(fmtMaint(a.Maintenance), px+10, y+62, FontXs, csTextDim)
		drawLabel("Size: "+a.Size, px+10, y+78, FontXs, csTextDim)
		drawLabel("Req: "+a.Requirements, px+10, y+94, FontXs, csTextDim)
		if a.UnlockPop > 0 {
			drawLabel(fmtUnlock(a.UnlockPop), px+10, y+110, FontXs, rl.NewColor(220, 190, 110, 255))
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
