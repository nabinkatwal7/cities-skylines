#!/usr/bin/env bash
# Fetch Kenney CC0 packs and install curated game assets.
# License: CC0 — https://kenney.nl
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ASSETS="$ROOT/assets"
TMP="${TMPDIR:-/tmp}/kenney-setup-$$"
mkdir -p "$TMP"

fetch() {
  local url="$1" out="$2"
  echo ">> $out"
  curl -fsSL -o "$TMP/$out" "$url"
}

fetch "https://kenney.nl/media/pages/assets/city-kit-commercial/16eb35d771-1753115042/kenney_city-kit-commercial_2.1.zip" commercial.zip
fetch "https://kenney.nl/media/pages/assets/city-kit-suburban/167f6dbc31-1745479373/kenney_city-kit-suburban_20.zip" suburban.zip
fetch "https://opengameart.org/sites/default/files/kenney_city-kit-industrial_1.0.zip" industrial.zip
fetch "https://kenney.nl/media/pages/assets/car-kit/1a312ec241-1775131960/kenney_car-kit.zip" car.zip
fetch "https://kenney.nl/media/pages/assets/nature-kit/8334871c74-1677698939/kenney_nature-kit.zip" nature.zip
fetch "https://kenney.nl/media/pages/assets/city-kit-roads/74288c9459-1741864740/kenney_city-kit-roads.zip" roads.zip

echo ">> clearing old assets"
rm -rf "$ASSETS/building" "$ASSETS/tree" "$ASSETS/downloads" "$ASSETS/fonts"
rm -f "$ASSETS/grass.png" "$ASSETS/road.jpg" "$ASSETS/road.png" "$ASSETS/water.png"
mkdir -p "$ASSETS/textures" "$ASSETS/buildings" "$ASSETS/trees" "$ASSETS/vehicles"

copy_colormap() {
  local zip="$1" dest="$2"
  mkdir -p "$dest/Textures"
  unzip -joq "$TMP/$zip" "Models/OBJ format/Textures/colormap.png" -d "$dest/Textures" 2>/dev/null || true
}

extract_obj() {
  local zip="$1" dest="$2"
  shift 2
  mkdir -p "$dest"
  copy_colormap "$zip" "$dest"
  local i=0
  for name in "$@"; do
    unzip -joq "$TMP/$zip" "Models/OBJ format/${name}.obj" "Models/OBJ format/${name}.mtl" -d "$dest"
    mv -f "$dest/${name}.obj" "$dest/${i}.obj"
    mv -f "$dest/${name}.mtl" "$dest/${i}.mtl"
    i=$((i + 1))
  done
}

extract_obj commercial.zip "$ASSETS/buildings/commercial_low" \
  building-a building-b building-c building-d
extract_obj commercial.zip "$ASSETS/buildings/commercial_high" \
  building-skyscraper-a building-skyscraper-b building-skyscraper-c building-skyscraper-d
extract_obj commercial.zip "$ASSETS/buildings/office" \
  building-e building-f building-g building-h
extract_obj suburban.zip "$ASSETS/buildings/residential_low" \
  building-type-a building-type-b building-type-c building-type-h
extract_obj suburban.zip "$ASSETS/buildings/residential_high" \
  building-type-m building-type-n building-type-t building-type-u
extract_obj industrial.zip "$ASSETS/buildings/industrial" \
  building-a building-b building-c building-d building-e building-f

copy_colormap nature.zip "$ASSETS/trees"
for pair in oak:tree_oak pine:tree_pineRoundA birch:tree_default palm:tree_palm dead:tree_oak_fall; do
  key="${pair%%:*}"
  obj="${pair##*:}"
  unzip -joq "$TMP/nature.zip" "Models/OBJ format/${obj}.obj" "Models/OBJ format/${obj}.mtl" -d "$ASSETS/trees"
  mv -f "$ASSETS/trees/${obj}.obj" "$ASSETS/trees/${key}.obj"
  mv -f "$ASSETS/trees/${obj}.mtl" "$ASSETS/trees/${key}.mtl"
done

copy_colormap car.zip "$ASSETS/vehicles"
for pair in car:sedan bus:van truck:truck emergency:ambulance bike:hatchback-sports; do
  key="${pair%%:*}"
  obj="${pair##*:}"
  unzip -joq "$TMP/car.zip" "Models/OBJ format/${obj}.obj" "Models/OBJ format/${obj}.mtl" -d "$ASSETS/vehicles"
  mv -f "$ASSETS/vehicles/${obj}.obj" "$ASSETS/vehicles/${key}.obj"
  mv -f "$ASSETS/vehicles/${obj}.mtl" "$ASSETS/vehicles/${key}.mtl"
done

unzip -joq "$TMP/roads.zip" "Models/OBJ format/Textures/colormap.png" -d "$ASSETS/textures" 2>/dev/null || true

cat >"$ASSETS/ATTRIBUTION.txt" <<'EOF'
3D assets from Kenney (www.kenney.nl) — CC0 1.0 Universal.
Packs: City Kit (Commercial, Suburban, Industrial), Car Kit, Nature Kit, City Kit (Roads).
No attribution required; see https://creativecommons.org/publicdomain/zero/1.0/
EOF

mkdir -p "$ASSETS/textures"
(cd "$ROOT" && go run ./cmd/gentextures)
rm -rf "$TMP"
echo ">> assets ready under $ASSETS"
