// Package math oferece funções utilitárias de cálculo geográfico.
package math

import "math"

// HaversineKm calcula a distância em linha reta entre dois pontos, em quilômetros.
func HaversineKm(aLat, aLng, bLat, bLng float64) float64 {
	const raioTerraKM = 6371.0

	phi1 := aLat * math.Pi / 180
	phi2 := bLat * math.Pi / 180
	dPhi := (bLat - aLat) * math.Pi / 180
	dLambda := (bLng - aLng) * math.Pi / 180

	s := math.Sin(dPhi/2)*math.Sin(dPhi/2) +
		math.Cos(phi1)*math.Cos(phi2)*math.Sin(dLambda/2)*math.Sin(dLambda/2)

	return raioTerraKM * 2 * math.Asin(math.Sqrt(s))
}
