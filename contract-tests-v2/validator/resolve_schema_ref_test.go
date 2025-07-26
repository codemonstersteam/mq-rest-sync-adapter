package validator

import (
	"testing"

	"github.com/codemonsters/contract-tests-v2/parser"
	"github.com/stretchr/testify/assert"
)

// Helper functions for creating test specifications

func createSpecWithSchemas(schemas map[string]parser.Schema) *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		Components: &parser.Components{
			Schemas: schemas,
		},
	}
}

func createNilSpec() *parser.AsyncAPISpec {
	return nil
}

func createSpecWithoutComponents() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{}
}

func createSpecWithNilComponents() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		Components: nil,
	}
}

func createSpecWithoutSchemas() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		Components: &parser.Components{
			Schemas: nil,
		},
	}
}

func createSpecWithEmptySchemas() *parser.AsyncAPISpec {
	return &parser.AsyncAPISpec{
		Components: &parser.Components{
			Schemas: map[string]parser.Schema{},
		},
	}
}

func createUserSchema() parser.Schema {
	return parser.Schema{
		Type: "object",
		Properties: map[string]parser.Schema{
			"id":   {Type: "integer"},
			"name": {Type: "string"},
		},
		Required: []string{"id", "name"},
	}
}

// createSchemaWithRef already defined in convert_schema_test.go

func createRecursiveSchemas() map[string]parser.Schema {
	return map[string]parser.Schema{
		"User": {
			Ref: "#/components/schemas/Person",
		},
		"Person": {
			Type: "object",
			Properties: map[string]parser.Schema{
				"id":   {Type: "integer"},
				"name": {Type: "string"},
			},
			Required: []string{"id", "name"},
		},
	}
}

func createCyclicSchemas() map[string]parser.Schema {
	return map[string]parser.Schema{
		"A": {
			Ref: "#/components/schemas/B",
		},
		"B": {
			Ref: "#/components/schemas/A", // Циклическая ссылка
		},
	}
}

// TestResolveSchemaRef тестирует основные сценарии согласно плану README.md
func TestResolveSchemaRef(t *testing.T) {
	v := NewChannelValidator()

	tests := []struct {
		name     string
		spec     *parser.AsyncAPISpec
		ref      string
		expected *parser.Schema
	}{
		// Основные тест-кейсы из плана README.md
		{
			name:     "валидная ссылка #/components/schemas/UserSchema",
			spec:     createSpecWithSchemas(map[string]parser.Schema{"UserSchema": createUserSchema()}),
			ref:      "#/components/schemas/UserSchema",
			expected: &parser.Schema{
				Type: "object",
				Properties: map[string]parser.Schema{
					"id":   {Type: "integer"},
					"name": {Type: "string"},
				},
				Required: []string{"id", "name"},
			},
		},
		{
			name:     "невалидная ссылка (неправильный префикс)",
			spec:     createSpecWithSchemas(map[string]parser.Schema{"User": createUserSchema()}),
			ref:      "#/invalid/schemas/User",
			expected: nil,
		},
		{
			name:     "ссылка на несуществующую схему",
			spec:     createSpecWithSchemas(map[string]parser.Schema{"User": createUserSchema()}),
			ref:      "#/components/schemas/NonExistent",
			expected: nil,
		},
		{
			name:     "nil спецификация",
			spec:     createNilSpec(),
			ref:      "#/components/schemas/User",
			expected: nil,
		},
		{
			name:     "спецификация без components/schemas",
			spec:     createSpecWithoutComponents(),
			ref:      "#/components/schemas/User",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := v.resolveSchemaRef(tt.spec, tt.ref)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestResolveSchemaRef_RecursiveLinks тестирует рекурсивные ссылки (схема → схема)
func TestResolveSchemaRef_RecursiveLinks(t *testing.T) {
	v := NewChannelValidator()

	t.Run("рекурсивная ссылка (схема → схема)", func(t *testing.T) {
		// Схема User ссылается на Person, который является конкретной схемой
		spec := createSpecWithSchemas(createRecursiveSchemas())
		
		result := v.resolveSchemaRef(spec, "#/components/schemas/User")
		
		expected := &parser.Schema{
			Type: "object",
			Properties: map[string]parser.Schema{
				"id":   {Type: "integer"},
				"name": {Type: "string"},
			},
			Required: []string{"id", "name"},
		}
		
		assert.Equal(t, expected, result)
	})

	t.Run("многоуровневая рекурсивная ссылка", func(t *testing.T) {
		// A -> B -> C (конкретная схема)
		schemas := map[string]parser.Schema{
			"A": {Ref: "#/components/schemas/B"},
			"B": {Ref: "#/components/schemas/C"},
			"C": {
				Type: "string",
				Enum: []interface{}{"value1", "value2"},
			},
		}
		spec := createSpecWithSchemas(schemas)
		
		result := v.resolveSchemaRef(spec, "#/components/schemas/A")
		
		expected := &parser.Schema{
			Type: "string",
			Enum: []interface{}{"value1", "value2"},
		}
		
		assert.Equal(t, expected, result)
	})

	t.Run("рекурсивная ссылка с битой цепочкой", func(t *testing.T) {
		// A -> B -> NonExistent
		schemas := map[string]parser.Schema{
			"A": {Ref: "#/components/schemas/B"},
			"B": {Ref: "#/components/schemas/NonExistent"}, // Битая ссылка
		}
		spec := createSpecWithSchemas(schemas)
		
		result := v.resolveSchemaRef(spec, "#/components/schemas/A")
		
		// Должно вернуть nil, так как цепочка разрешения прерывается
		assert.Nil(t, result)
	})
}

// TestResolveSchemaRef_CyclicLinks тестирует циклические ссылки (A → B → A)
func TestResolveSchemaRef_CyclicLinks(t *testing.T) {
	v := NewChannelValidator()

	t.Run("циклическая ссылка (A → B → A)", func(t *testing.T) {
		spec := createSpecWithSchemas(createCyclicSchemas())
		
		// Теперь функция защищена от циклических ссылок
		result := v.resolveSchemaRef(spec, "#/components/schemas/A")
		
		// Циклические ссылки должны возвращать nil (защита от бесконечной рекурсии)
		assert.Nil(t, result)
		t.Logf("Циклическая ссылка корректно обработана: %v", result)
	})

	t.Run("прямая циклическая ссылка A→A", func(t *testing.T) {
		schemas := map[string]parser.Schema{
			"A": {Ref: "#/components/schemas/A"}, // Прямая циклическая ссылка
		}
		spec := createSpecWithSchemas(schemas)
		
		result := v.resolveSchemaRef(spec, "#/components/schemas/A")
		
		// Прямая циклическая ссылка должна возвращать nil
		assert.Nil(t, result)
		t.Logf("Прямая циклическая ссылка корректно обработана: %v", result)
	})
}

// TestResolveSchemaRef_EdgeCases тестирует дополнительные полезные случаи
func TestResolveSchemaRef_EdgeCases(t *testing.T) {
	v := NewChannelValidator()

	t.Run("спецификация с nil components", func(t *testing.T) {
		spec := createSpecWithNilComponents()
		result := v.resolveSchemaRef(spec, "#/components/schemas/User")
		assert.Nil(t, result)
	})

	t.Run("спецификация без schemas", func(t *testing.T) {
		spec := createSpecWithoutSchemas()
		result := v.resolveSchemaRef(spec, "#/components/schemas/User")
		assert.Nil(t, result)
	})

	t.Run("спецификация с пустой schemas map", func(t *testing.T) {
		spec := createSpecWithEmptySchemas()
		result := v.resolveSchemaRef(spec, "#/components/schemas/User")
		assert.Nil(t, result)
	})

	t.Run("пустая ссылка", func(t *testing.T) {
		spec := createSpecWithSchemas(map[string]parser.Schema{"User": createUserSchema()})
		result := v.resolveSchemaRef(spec, "")
		assert.Nil(t, result)
	})

	t.Run("ссылка с только префиксом", func(t *testing.T) {
		spec := createSpecWithSchemas(map[string]parser.Schema{"User": createUserSchema()})
		result := v.resolveSchemaRef(spec, "#/components/schemas/")
		assert.Nil(t, result)
	})

	t.Run("case sensitive имена схем", func(t *testing.T) {
		schemas := map[string]parser.Schema{
			"User": createUserSchema(),
		}
		spec := createSpecWithSchemas(schemas)
		
		// Неправильный регистр должен вернуть nil
		result := v.resolveSchemaRef(spec, "#/components/schemas/user")
		assert.Nil(t, result)
		
		// Правильный регистр должен работать
		result = v.resolveSchemaRef(spec, "#/components/schemas/User")
		assert.NotNil(t, result)
	})

	t.Run("ссылка с дополнительными слэшами", func(t *testing.T) {
		schemas := map[string]parser.Schema{
			"User": createUserSchema(),
		}
		spec := createSpecWithSchemas(schemas)
		
		result := v.resolveSchemaRef(spec, "#/components/schemas//User")
		
		// Должно вернуть nil, так как имя схемы "/User" недопустимо по AsyncAPI 3.0
		assert.Nil(t, result)
	})
}

// TestResolveSchemaRef_AsyncAPI30_SchemaNames тестирует валидацию имен схем согласно AsyncAPI 3.0
func TestResolveSchemaRef_AsyncAPI30_SchemaNames(t *testing.T) {
	v := NewChannelValidator()

	validNames := []struct {
		name       string
		schemaName string
	}{
		{"простое имя", "User"},
		{"имя с цифрами", "User123"},
		{"имя с дефисом", "User-Profile"},
		{"имя с подчеркиванием", "User_Profile"},
		{"имя с точкой", "User.Profile"},
		{"комбинация всех допустимых символов", "User123-Profile_v2.0"},
		{"только цифры", "123"},
		{"только символы", "ABC"},
		{"начинается с цифры", "1User"},
		{"начинается с дефиса", "-User"},
		{"начинается с точки", ".User"},
	}

	for _, tt := range validNames {
		t.Run("валидное имя: "+tt.name, func(t *testing.T) {
			schemas := map[string]parser.Schema{
				tt.schemaName: createUserSchema(),
			}
			spec := createSpecWithSchemas(schemas)
			
			result := v.resolveSchemaRef(spec, "#/components/schemas/"+tt.schemaName)
			
			assert.NotNil(t, result, "Имя схемы '%s' должно быть валидным по AsyncAPI 3.0", tt.schemaName)
			assert.Equal(t, "object", result.Type)
		})
	}

	invalidNames := []struct {
		name       string
		schemaName string
		reason     string
	}{
		{"пустое имя", "", "пустое имя недопустимо"},
		{"имя с пробелом", "User Profile", "пробелы недопустимы"},
		{"имя со слэшем", "User/Profile", "слэши недопустимы"},
		{"имя с @", "User@Profile", "@ недопустим"},
		{"имя с #", "User#Profile", "# недопустим"},
		{"имя с $", "User$Profile", "$ недопустим"},
		{"имя с %", "User%Profile", "% недопустим"},
		{"имя с &", "User&Profile", "& недопустим"},
		{"имя с *", "User*Profile", "* недопустим"},
		{"имя с +", "User+Profile", "+ недопустим"},
		{"имя с =", "User=Profile", "= недопустим"},
		{"имя с скобками", "User(Profile)", "скобки недопустимы"},
		{"имя с квадратными скобками", "User[Profile]", "квадратные скобки недопустимы"},
		{"имя с фигурными скобками", "User{Profile}", "фигурные скобки недопустимы"},
		{"имя с unicode", "Пользователь", "unicode символы недопустимы"},
	}

	for _, tt := range invalidNames {
		t.Run("невалидное имя: "+tt.name, func(t *testing.T) {
			schemas := map[string]parser.Schema{
				tt.schemaName: createUserSchema(),
			}
			spec := createSpecWithSchemas(schemas)
			
			result := v.resolveSchemaRef(spec, "#/components/schemas/"+tt.schemaName)
			
			assert.Nil(t, result, "Имя схемы '%s' должно быть невалидным по AsyncAPI 3.0: %s", tt.schemaName, tt.reason)
		})
	}
}

// TestResolveSchemaRef_InvalidPrefixes тестирует различные невалидные префиксы
func TestResolveSchemaRef_InvalidPrefixes(t *testing.T) {
	v := NewChannelValidator()
	
	spec := createSpecWithSchemas(map[string]parser.Schema{"User": createUserSchema()})

	invalidRefs := []struct {
		name string
		ref  string
	}{
		{"без префикса", "User"},
		{"неправильный префикс", "#/invalid/schemas/User"},
		{"частичный префикс", "#/components/User"},
		{"ссылка на messages", "#/components/messages/User"},
		{"ссылка на servers", "#/components/servers/User"},
		{"без components", "#/schemas/User"},
		{"пустой префикс", "#/User"},
		{"только решетка", "#User"},
		{"без решетки", "/components/schemas/User"},
	}

	for _, tt := range invalidRefs {
		t.Run(tt.name, func(t *testing.T) {
			result := v.resolveSchemaRef(spec, tt.ref)
			assert.Nil(t, result, "Невалидная ссылка %s должна возвращать nil", tt.ref)
		})
	}
}

// TestResolveSchemaRef_BugRegression тестирует регрессию потенциальных багов
func TestResolveSchemaRef_BugRegression(t *testing.T) {
	v := NewChannelValidator()

	t.Run("REGRESSION: схема только с ref полем должна резолвиться", func(t *testing.T) {
		// Важный кейс: схема содержит только Ref поле
		schemas := map[string]parser.Schema{
			"UserAlias": {Ref: "#/components/schemas/User"},
			"User": {
				Type: "object",
				Properties: map[string]parser.Schema{
					"id": {Type: "string"},
				},
			},
		}
		spec := createSpecWithSchemas(schemas)
		
		result := v.resolveSchemaRef(spec, "#/components/schemas/UserAlias")
		
		assert.NotNil(t, result)
		assert.Equal(t, "object", result.Type)
		assert.Contains(t, result.Properties, "id")
	})

	t.Run("REGRESSION: схема с ref и другими полями", func(t *testing.T) {
		// Тест что ref имеет приоритет над другими полями
		schemas := map[string]parser.Schema{
			"UserAlias": {
				Ref:  "#/components/schemas/User",
				Type: "string", // Должно игнорироваться из-за Ref
			},
			"User": {
				Type: "object",
				Properties: map[string]parser.Schema{
					"name": {Type: "string"},
				},
			},
		}
		spec := createSpecWithSchemas(schemas)
		
		result := v.resolveSchemaRef(spec, "#/components/schemas/UserAlias")
		
		assert.NotNil(t, result)
		assert.Equal(t, "object", result.Type) // Должно быть object, не string
		assert.Contains(t, result.Properties, "name")
	})

	t.Run("REGRESSION: пустое имя схемы после TrimPrefix", func(t *testing.T) {
		// Тест для случая, когда TrimPrefix может дать пустую строку
		schemas := map[string]parser.Schema{
			"": createUserSchema(), // Пустое имя схемы (недопустимо по AsyncAPI 3.0)
		}
		spec := createSpecWithSchemas(schemas)
		
		result := v.resolveSchemaRef(spec, "#/components/schemas/")
		
		// Пустые имена схем недопустимы согласно AsyncAPI 3.0 спецификации
		assert.Nil(t, result)
	})
}