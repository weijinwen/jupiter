package meilisearch

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type UserFilter struct {
	Name     string    `filter:"name,op=CONTAINS"`
	Age      int       `filter:"age,op=>="`
	Status   string    `filter:"status,op=IN,required"`
	Email    string    `filter:"email,ignorezero"`
	JoinDate time.Time `filter:"join_date,op=>=,datefmt=2006-01-02"`
	Tags     []string  `filter:"tags,op=CONTAINS"`
	IsActive bool      `filter:"is_active,op=="`
	Location Location  `filter:"location,nested"`
}

type Location struct {
	City    string `filter:"city"`
	Country string `filter:"country,required"`
}

type ProductFilter struct {
	Name       string            `filter:"name,op=CONTAINS"`
	Category   string            `filter:"category,required"`
	Price      float64           `filter:"price,op=>="`
	InStock    bool              `filter:"in_stock,op=="`
	CreatedAt  time.Time         `filter:"created_at,op=>=,datefmt=2006-01-02"`
	Tags       []string          `filter:"tags,op=CONTAINS"`
	Attributes ProductAttributes `filter:"attributes,nested"`
}

type ProductAttributes struct {
	Brand    string `filter:"brand,required"`
	Color    string `filter:"color"`
	Size     string `filter:"size,ignorezero"`
	Material string `filter:"material"`
}

func ExampleUsage() {
	builder := NewBuilder()

	builder.RegisterValidator("age", func(value any) error {
		if age, ok := value.(int); ok {
			if age < 0 || age > 150 {
				return errors.New("age must be between 0 and 150")
			}
		}
		return nil
	})

	builder.RegisterTransformer("email", func(value any) any {
		if email, ok := value.(string); ok {
			return strings.ToLower(email)
		}
		return value
	})

	userFilter := UserFilter{
		Name:     "John",
		Age:      25,
		Status:   "active",
		Email:    "John.Doe@Example.com",
		JoinDate: time.Now().AddDate(-1, 0, 0),
		Tags:     []string{"developer", "golang"},
		IsActive: true,
		Location: Location{
			City:    "New York",
			Country: "USA",
		},
	}

	builder.AddFilterStruct(userFilter)

	geoFilter := GeoFilter{
		Point:  GeoPoint{Lat: 40.7128, Lng: -74.0060},
		Radius: 10000,
		Unit:   "m",
	}
	builder.AddGeoFilter(geoFilter)

	dateFilter := DateRangeFilter{
		Field: "created_at",
		Start: time.Now().AddDate(-1, 0, 0),
		End:   time.Now(),
	}
	builder.AddDateRangeFilter(dateFilter)

	arrayFilter := ArrayFilter{
		Field:    "categories",
		Values:   []any{"electronics", "books"},
		Operator: "IN",
	}
	builder.AddArrayFilter(arrayFilter)

	productFilter := ProductFilter{
		Name:      "Smartphone",
		Category:  "electronics",
		Price:     299.99,
		InStock:   true,
		CreatedAt: time.Now().AddDate(-2, 0, 0),
		Tags:      []string{"smartphone", "android"},
		Attributes: ProductAttributes{
			Brand:    "TechCorp",
			Color:    "black",
			Size:     "6.5 inch",
			Material: "glass",
		},
	}

	builder.AddFilterStructWithGroup("product", productFilter)

	condition := FilterCondition{
		Filters: []Filter{
			{Field: "rating", Operator: ">=", Value: 4},
			{Field: "reviews", Operator: ">", Value: 10},
		},
		Operator: "AND",
	}
	builder.AddFilterGroup("product", condition)

	builder.ApplyFilterGroup("product")

	builder.SetQuery("search query").
		SetPage(1).
		SetHitsPerPage(20).
		AddSort("price", false).
		AddFacets("category", "brand").
		AddHighlight("name", "description").
		EnableMatches().
		ShowRankingScore()

	searchRequest := builder.Build()

	fmt.Printf("Generated search request: %+v\n", searchRequest)
}

func ExampleAdvancedFiltering() {
	builder := NewBuilder()

	orCondition := FilterCondition{
		Filters: []Filter{
			{Field: "status", Operator: "=", Value: "active"},
			{Field: "status", Operator: "=", Value: "pending"},
		},
		Operator: "OR",
	}
	builder.AddFilterCondition(orCondition)

	andCondition := FilterCondition{
		Filters: []Filter{
			{Field: "age", Operator: ">=", Value: 18},
			{Field: "verified", Operator: "=", Value: true},
		},
		Operator: "AND",
	}
	builder.AddFilterCondition(andCondition)

	builder.AddFilter("email", "CONTAINS", "gmail.com")

	searchRequest := builder.Build()

	fmt.Printf("Advanced filter request: %+v\n", searchRequest)
}

func ExampleNestedStructFiltering() {
	type ComplexFilter struct {
		UserInfo UserFilter    `filter:"user_info,nested"`
		Product  ProductFilter `filter:"product,nested"`
		Price    float64       `filter:"price,op=>=,required"`
		Date     time.Time     `filter:"date,op=>=,datefmt=2006-01-02,required"`
	}

	builder := NewBuilder()

	complexFilter := ComplexFilter{
		UserInfo: UserFilter{
			Name:     "Alice",
			Age:      30,
			Status:   "active",
			Location: Location{City: "Boston", Country: "USA"},
		},
		Product: ProductFilter{
			Name:       "Laptop",
			Category:   "electronics",
			Price:      999.99,
			InStock:    true,
			Attributes: ProductAttributes{Brand: "TechBrand", Color: "silver"},
		},
		Price: 500.00,
		Date:  time.Now().AddDate(-1, 0, 0),
	}

	builder.AddFilterStruct(complexFilter)

	searchRequest := builder.Build()

	fmt.Printf("Nested struct filter request: %+v\n", searchRequest)
}

func ExampleTypeSafeArrayFilter() {
	builder := NewBuilder()

	builder.AddFilter("tags", "IN", []string{"golang", "rust", "python"})

	builder.AddFilter("ids", "IN", []int{1, 2, 3, 4, 5})

	builder.AddFilter("prices", "IN", []float64{9.99, 19.99, 29.99})

	searchRequest := builder.Build()

	fmt.Printf("Type-safe array filter request: %+v\n", searchRequest)
}

func ExampleClientUsage() {
	fmt.Println("Client usage example (requires actual MeiliSearch server):")
	fmt.Println(`
    // 创建客户端
    client, err := meilisearch.StdConfig("default").Singleton()
    if err != nil {
        panic(err)
    }

    // 执行搜索
    resp, err := client.Search("products", &meilisearch.SearchRequest{
        Query: "smartphone",
        Page:  1,
        HitsPerPage: 20,
    })

    // 获取文档
    var result meilisearch.DocumentsResult
    err = client.GetDocuments("products", &meilisearch.DocumentsQuery{
        Limit: 10,
    }, &result)

    // 添加文档
    docs := []map[string]any{
        {"id": 1, "name": "Product 1", "price": 99.99},
        {"id": 2, "name": "Product 2", "price": 149.99},
    }
    task, err := client.AddDocuments("products", docs, nil)

    // 更新文档
    task, err = client.UpdateDocuments("products", docs, nil)

    // 删除文档
    task, err = client.DeleteDocument("products", "1")

    // 删除所有文档
    task, err = client.DeleteAllDocuments("products")
    `)
}
