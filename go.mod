module github.com/ZeroMarker/clashon

go 1.24

require github.com/metacubex/mihomo v1.19.32

replace github.com/metacubex/mihomo => ./core/mihomo

// Mirrors core/mihomo/go.mod; replace directives are not inherited.
replace google.golang.org/protobuf => github.com/metacubex/protobuf-go v0.0.0-20260306035419-7ceee0674686
