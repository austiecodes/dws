package schema

type Image struct {
	ID         uint   `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UUID       int    `gorm:"column:uuid" json:"uuid"`
	ImageID    string `gorm:"column:image_id" json:"image_id"`
	Repository string `gorm:"column:repository" json:"repository"`
	Tag        string `gorm:"column:tag" json:"tag"`
	Created    string `gorm:"column:created" json:"created"`
	Size       string `gorm:"column:size" json:"size"`
}
