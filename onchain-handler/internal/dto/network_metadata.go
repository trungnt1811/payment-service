package dto

type NetworkMetadataDTO struct {
	ID         uint64 `json:"id"`
	Network    string `json:"network"`
	IconBase64 string `json:"icon_base64"`
}
