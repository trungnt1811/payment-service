package dto

type NetworkMetadataDTO struct {
	ID         uint64 `json:"id"`
	Alias      string `json:"alias"`
	Name       string `json:"name"`
	IconBase64 string `json:"icon_base64"`
}
