package gohomework3

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"gorm.io/gorm"
)

// 题目1：模型定义
// 假设你要开发一个博客系统，有以下几个实体： User （用户）、 Post （文章）、 Comment （评论）。
// 要求 ：
// 使用Gorm定义 User 、 Post 和 Comment 模型，其中 User 与 Post 是一对多关系（一个用户可以发布多篇文章）， Post 与 Comment 也是一对多关系（一篇文章可以有多个评论）。
// 编写Go代码，使用Gorm创建这些模型对应的数据库表。
// 题目2：关联查询
// 基于上述博客系统的模型定义。
// 要求 ：
// 编写Go代码，使用Gorm查询某个用户发布的所有文章及其对应的评论信息。
// 编写Go代码，使用Gorm查询评论数量最多的文章信息。
// 题目3：钩子函数
// 继续使用博客系统的模型。
// 要求 ：
// 为 Post 模型添加一个钩子函数，在文章创建时自动更新用户的文章数量统计字段。
// 为 Comment 模型添加一个钩子函数，在评论删除时检查文章的评论数量，如果评论数量为 0，则更新文章的评论状态为 "无评论"。
type User struct {
	ID        uint
	Name      string
	Age       uint8
	Email     string
	PostCount uint
	Posts     []Post
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Post struct {
	ID            uint
	Title         string
	Content       string
	UserID        uint
	CommentCount  uint
	CommentStatus string
	Comments      []Comment
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Comment struct {
	ID        uint
	Content   string
	PostID    uint
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (p *Post) AfterCreate(tx *gorm.DB) error {
	var userData User
	if err := tx.Where("id = ?", p.UserID).Find(&userData).Error; err != nil {
		return err
	}
	if err := tx.Model(&User{}).Where("id = ?", p.UserID).Update("post_count", gorm.Expr("post_count + ?", 1)).Error; err != nil {
		return err
	}
	return nil
}

func (c *Comment) AfterCreate(tx *gorm.DB) error {
	var postData Post
	if err := tx.Where("id = ?", c.PostID).Find(&postData).Error; err != nil {
		return err
	}
	if err := tx.Model(&Post{}).Where("id = ?", c.PostID).Update("comment_count", gorm.Expr("comment_count + ?", 1)).Error; err != nil {
		return err
	}
	return nil
}

func (c *Comment) AfterDelete(tx *gorm.DB) error {
	fmt.Println("删除前值：", c)
	var comm Comment
	if err := tx.Where("id = ?", c.ID).Find(&comm).Error; err != nil {
		return err
	}
	fmt.Println("查询值：", comm)

	var commentCount int64
	if err := tx.Model(&Post{}).Where("id = ?", c.PostID).Update("comment_count", gorm.Expr("comment_count - ?", 1)).Error; err != nil {
		return err
	}
	if err := tx.Model(&Comment{}).Where("post_id = ?", c.PostID).Count(&commentCount).Error; err != nil {
		return err
	}
	if commentCount == 0 {
		if err := tx.Model(&Post{}).Where("id = ?", c.PostID).Update("Comment_Status", "无评论").Error; err != nil {
			return err
		}
	}
	return nil
}

func TestAutoMigrate(t *testing.T) {
	db := NewTestDB(t, "blog.db")
	if err := db.AutoMigrate(&User{}, &Post{}, &Comment{}); err != nil {
		t.Fatalf("auto migrate:%v", err)
	}
}

func TestCreateData(t *testing.T) {
	db := NewTestDB(t, "blog.db")
	data := User{
		Name:  "忘语",
		Age:   49,
		Email: "wangyu@test.com",
		Posts: []Post{
			{Title: "凡人修仙传人界篇",
				Content: "韩立在人界的故事",
				Comments: []Comment{
					{Content: "人界篇夯爆了"},
					{Content: "去SPA"},
					{Content: "南宫婉掩月宗"},
					{Content: "吕落千浪决"},
				},
			},
			{Title: "凡人修仙传灵界篇",
				Content: "韩立飞升灵界的故事",
				Comments: []Comment{
					{Content: "向之礼"},
					{Content: "灵界马良夯爆了"},
					{Content: "阳鹿真灵"},
					{Content: "宝花智慧与美貌共存"},
					{Content: "玄天斩灵剑"},
				},
			},
			{Title: "凡人修仙传仙界篇",
				Content: "韩立飞升仙界的故事",
				Comments: []Comment{
					{Content: "柳石哥哥"},
					{Content: "本源道祖"},
					{Content: "轮回殿主"},
					{Content: "古或今"},
					{Content: "元瑶"},
					{Content: "青元子"},
				},
			},
		},
	}
	if err := db.Session(&gorm.Session{FullSaveAssociations: true}).Create(&data).Error; err != nil {
		t.Fatalf("create data:%v", err)
	}
}

func TestQueryData(t *testing.T) {
	db := NewTestDB(t, "blog.db")
	var userData User
	if err := db.Preload("Posts").Preload("Posts.Comments").Where("id = ?", 1).Find(&userData).Error; err != nil {
		t.Fatalf("query user:%v", err)
	}
	jsonData, err := json.MarshalIndent(userData, "", " ")
	if err != nil {
		t.Fatalf("转json失败 %v", err)
	}

	fmt.Println("查询到的数据", string(jsonData))

}
func TestQueryMaxComment(t *testing.T) {
	db := NewTestDB(t, "blog.db")

	var postData Post
	if err := db.Model(&Post{}).Preload("Comments").Order("comment_count desc").First(&postData).Error; err != nil {
		t.Fatalf("CommnetCount max first:%v", err)
	}

	postJson, err := json.MarshalIndent(postData, "", " ")
	if err != nil {
		t.Fatalf("转文章json失败 %v", err)
	}
	fmt.Println("评论最大的文章", string(postJson))

	var postDataSql Post
	if err := db.Raw(`
		select a.*,count(b.id) count from posts a left join comments b on b.post_id = a.id group by b.post_id order by count desc limit 1
	`).Scan(&postDataSql).Error; err != nil {
		t.Fatalf("sql查询最大评论的文章：%v", err)
	}

	postJsonSql, err := json.MarshalIndent(postDataSql, "", " ")
	if err != nil {
		t.Fatalf("转文章json失败 %v", err)
	}
	fmt.Println("评论最大的文章", string(postJsonSql))
}

func TestInsertPost(t *testing.T) {
	db := NewTestDB(t, "blog.db")
	postData := Post{
		Title:   "凡人修仙传外传",
		Content: "黑历史",
		UserID:  1,
	}

	var userBeforeCount uint
	if err := db.Model(&User{}).Select("post_count").Where("id = ?", 1).Find(&userBeforeCount).Error; err != nil {
		t.Fatalf("query user post count:%v", err)
	}
	fmt.Printf("用户之前的文章数量：%d\n", userBeforeCount)

	if err := db.Model(&Post{}).Create(&postData).Error; err != nil {
		t.Fatalf("create post:%v", err)
	}

	var userData User
	if err := db.Model(&User{}).Where("id = ?", 1).First(&userData).Error; err != nil {
		t.Fatalf("query user:%v", err)
	}
	if userData.PostCount-userBeforeCount != 1 {
		fmt.Printf("新增文章后数量没有增加 %d - %d\n", userData.PostCount, userBeforeCount)
	}
}

func TestDeleteComment(t *testing.T) {
	db := NewTestDB(t, "blog.db")

	commentId := []int{1, 2, 3, 4}
	for _, i := range commentId {

		var c Comment
		if err := db.Where("id = ?", i).Find(&c).Error; err != nil {
			t.Fatalf("query comment:%v", err)
		}

		var posts Post
		if err := db.Where("id = ?", c.PostID).Find(&posts).Error; err != nil {
			t.Fatalf("before query post :%v", err)
		}
		fmt.Println("查询前", posts)
		if err := db.Delete(&c).Error; err != nil {
			t.Fatalf("delete comment %v", err)
		}
		var postAfter Post
		if err := db.Where("id = ?", c.PostID).Find(&postAfter).Error; err != nil {
			t.Fatalf("after query post:%v", err)
		}
		fmt.Println("查询删除后：", postAfter)

	}

}
