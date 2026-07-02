package ui

import (
	"fmt"
	"strings"

	"github.com/katwate/js-skylines/internal/road"
	"github.com/katwate/js-skylines/internal/sim"
	"github.com/katwate/js-skylines/internal/transport"
	"github.com/katwate/js-skylines/internal/zoning"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type SearchKind int

const (
	SearchBuilding SearchKind = iota
	SearchStreet
	SearchDistrict
	SearchTransit
	SearchCitizen
	SearchService
)

// SearchResult is one entity match (24.16).
type SearchResult struct {
	Kind  SearchKind
	Label string
	X, Z  float32
}

// SearchSystem finds entities and centers the camera (24.16).
type SearchSystem struct {
	open     bool
	persist  bool
	query    string
	results  []SearchResult
	selected int
	camX     float32
	camZ     float32
	hasCam   bool
}

func NewSearchSystem() *SearchSystem { return &SearchSystem{} }

func (s *SearchSystem) Open()  { s.open = true }
func (s *SearchSystem) Close() { s.open = false; s.query = "" }

func (s *SearchSystem) IsOpen() bool { return s.open }

func (s *SearchSystem) CameraTarget() (x, z float32, ok bool) {
	if !s.hasCam {
		return 0, 0, false
	}
	return s.camX, s.camZ, true
}

func (s *SearchSystem) ClearCamera() { s.hasCam = false }

func (s *SearchSystem) HandleInput() {
	if !s.open {
		return
	}
	if rl.IsKeyPressed(rl.KeyEscape) {
		s.Close()
	}
	if rl.IsKeyPressed(rl.KeyEnter) && len(s.results) > 0 {
		r := s.results[s.selected]
		s.camX, s.camZ = r.X, r.Z
		s.hasCam = true
		s.open = false
	}
	if rl.IsKeyPressed(rl.KeyDown) && len(s.results) > 0 {
		s.selected = (s.selected + 1) % len(s.results)
	}
	if rl.IsKeyPressed(rl.KeyUp) && len(s.results) > 0 {
		s.selected--
		if s.selected < 0 {
			s.selected = len(s.results) - 1
		}
	}
	ch := rl.GetCharPressed()
	for ch != 0 {
		if ch == 8 && len(s.query) > 0 {
			s.query = s.query[:len(s.query)-1]
		} else if ch >= 32 && ch < 127 {
			s.query += string(rune(ch))
		}
		ch = rl.GetCharPressed()
	}
}

func (s *SearchSystem) Update(sm *sim.SimulationManager) {
	s.results = s.results[:0]
	s.selected = 0
	q := strings.ToLower(strings.TrimSpace(s.query))
	if q == "" || sm == nil {
		return
	}
	if sm.Buildings != nil {
		for i := range sm.Buildings.Buildings {
			b := &sm.Buildings.Buildings[i]
			name := strings.ToLower(zoneTypeName(b.Type))
			if strings.Contains(name, q) || strings.Contains(fmt.Sprintf("b%d", b.ID), q) {
				s.results = append(s.results, SearchResult{
					Kind: SearchBuilding, Label: zoneTypeName(b.Type),
					X: b.WorldX, Z: b.WorldZ,
				})
			}
		}
	}
	if sm.Roads != nil {
		for i := range sm.Roads.Segments {
			seg := sm.Roads.Segments[i]
			rt := strings.ToLower(road.RoadTypeName(seg.RoadType))
			if strings.Contains(rt, q) || strings.Contains("road", q) {
				na := &sm.Roads.Nodes[seg.NodeA]
				nb := &sm.Roads.Nodes[seg.NodeB]
				s.results = append(s.results, SearchResult{
					Kind: SearchStreet, Label: road.RoadTypeName(seg.RoadType) + " Rd",
					X: (na.X + nb.X) * 0.5, Z: (na.Z + nb.Z) * 0.5,
				})
			}
		}
	}
	if sm.Transport != nil {
		for _, line := range sm.Transport.Lines {
			if strings.Contains(strings.ToLower(line.Name), q) {
				if len(line.Stops) > 0 {
					if st := sm.Transport.StopByID(line.Stops[0]); st != nil {
						s.results = append(s.results, SearchResult{
							Kind: SearchTransit, Label: line.Name,
							X: st.X, Z: st.Z,
						})
					}
				}
			}
		}
		for _, stop := range sm.Transport.Stops {
			if strings.Contains(strings.ToLower(transport.TypeName(stop.TransType)), q) {
				s.results = append(s.results, SearchResult{
					Kind: SearchService, Label: transport.TypeName(stop.TransType) + " stop",
					X: stop.X, Z: stop.Z,
				})
			}
		}
	}
	if sm.Zones != nil && strings.Contains(q, "district") {
		s.results = append(s.results, SearchResult{
			Kind: SearchDistrict, Label: "District 1", X: 0, Z: 0,
		})
	}
	if strings.Contains(q, "citizen") && sm.Buildings != nil {
		for i := range sm.Buildings.Buildings {
			b := &sm.Buildings.Buildings[i]
			if zoning.ZoneCategoryOf(b.Type) == zoning.CategoryResidential {
				s.results = append(s.results, SearchResult{
					Kind: SearchCitizen, Label: fmt.Sprintf("Citizen near %s", zoneTypeName(b.Type)),
					X: b.WorldX, Z: b.WorldZ,
				})
				break
			}
		}
	}
	if len(s.results) > 12 {
		s.results = s.results[:12]
	}
}

func (s *SearchSystem) Draw() {
	if !s.open {
		return
	}
	w, h := int32(440), int32(280)
	x := int32((ScreenW - w) / 2)
	y := int32(TopBarH + 48)
	drawPanel(x, y, w, h)
	drawLabel(T("search.title"), x+14, y+12, FontLg, csText)
	csInputField(x+14, y+40, w-28, 30)
	q := s.query
	if q == "" {
		q = T("search.placeholder")
	}
	drawLabel(q, x+20, y+46, FontMd, csTextDim)
	for i, r := range s.results {
		col := csTextDim
		if i == s.selected {
			col = csBarLine
		}
		drawLabel(fmt.Sprintf("%s — %s", searchKindName(r.Kind), r.Label), x+14, y+82+int32(i*22), FontMd, col)
	}
}

func searchKindName(k SearchKind) string {
	switch k {
	case SearchBuilding:
		return "Building"
	case SearchStreet:
		return "Street"
	case SearchDistrict:
		return "District"
	case SearchTransit:
		return "Transit"
	case SearchCitizen:
		return "Citizen"
	case SearchService:
		return "Service"
	default:
		return "?"
	}
}
