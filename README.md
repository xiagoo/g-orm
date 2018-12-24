需要定义类似结构体

type Model struct {
	master *gorm.DB
	slave  *gorm.DB
	stats  *gorm.DB
}