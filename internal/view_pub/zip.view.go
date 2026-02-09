package view_pub

type Zip struct {
	ID           string  `db:"id" json:"id,omitempty"`
	ProvinceId   *string `db:"province_id" json:"province_id,omitempty"`
	ProvinceName *string `db:"province_name" json:"province_name,omitempty"`
	CityId       *string `db:"city_id" json:"city_id,omitempty"`
	CityType     *string `db:"city_type" json:"city_type,omitempty"`
	CityName     *string `db:"city_name" json:"city_name,omitempty"`
	KecId        *string `db:"kec_id" json:"kec_id,omitempty"`
	KecName      *string `db:"kec_name" json:"kec_name,omitempty"`
	KelId        *string `db:"kel_id" json:"kel_id,omitempty"`
	KelName      *string `db:"kel_name" json:"kel_name,omitempty"`
	Label        *string `db:"label" json:"label,omitempty"`
	Zip          *string `db:"zip" json:"zip,omitempty"`
}
