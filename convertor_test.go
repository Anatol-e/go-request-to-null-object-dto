package convertor

import (
	"database/sql"
	"testing"
)

type UserJsonRequest struct {
	Name        string   `json:"name"`
	Age         *int64   `json:"age"`
	IsActive    *bool    `json:"is_active"`
	Hg          *float64 `json:"hg"`
	SimpleFloat float32  `json:"simple_float"`
	SimpleInt   int      `json:"simple_int"`
}

type UserDTO struct {
	Name        sql.NullString  `dto:"name"`
	Age         sql.NullInt64   `dto:"age"`
	IsActive    sql.NullBool    `dto:"is_active"`
	Hg          sql.NullFloat64 `dto:"hg"`
	SimpleFloat sql.NullFloat64 `dto:"simple_float"`
	SimpleInt   sql.NullInt32   `dto:"simple_int"`
}

type testStruct struct {
	Val1   string               `json:"val"`
	Val2   *teststructneasted1  `json:"val_str"`
	ValArr []teststructneasted1 `json:"val_arr"`
}

type teststructneasted1 struct {
	Val1 string   `json:"val"`
	Val3 []string `json:"val_arr"`
}

type testStructDto struct {
	Val1   sql.NullString          `dto:"val"`
	Val2   *teststructneastedDto1  `dto:"val_str"`
	ValArr []teststructneastedDto1 `dto:"val_arr"`
}

type teststructneastedDto1 struct {
	Val1 sql.NullString   `dto:"val"`
	Val3 []sql.NullString `dto:"val_arr"`
}

func TestConvertorWhenStructureIdentic(t *testing.T) {
	age := int64(1)
	userRequest := UserJsonRequest{
		Name: "Anatol",
		Age:  &age,
	}
	userDto := UserDTO{}
	err := ConvertToDTO(userRequest, &userDto)
	if err != nil {
		t.Fatal(err)
	}
}

func TestConvertorWhenStructureNotIdenticByFieldsNumbers(t *testing.T) {
	userRequest := struct {
		Age  *int64 `json:"age"`
		Name string `json:"name"`
	}{
		Name: "Anatol",
	}
	userDto := struct {
		Name sql.NullString `dto:"name"`
	}{}
	err := ConvertToDTO(userRequest, &userDto)
	if err == nil {
		t.Fatal("should be number of fields are not identic error")
	}
}

func TestConvertorWhenStructureNotIdenticByTagsNumbers(t *testing.T) {
	userRequest := struct {
		Age  *int64 `json:"age"`
		Name string `json:"name"`
	}{
		Name: "Anatol",
	}
	userDto := struct {
		Age  sql.NullInt64  `dto:"ages"`
		Name sql.NullString `dto:"name"`
	}{}
	err := ConvertToDTO(userRequest, &userDto)
	if err == nil {
		t.Fatal("should be tags of fields are not identic error")
	}
}

func TestConvertorWhenStructureNotIdenticByTypes(t *testing.T) {
	userRequest := struct {
		Age  *int64 `json:"age"`
		Name string `json:"name"`
	}{
		Name: "Anatol",
	}
	userDto := struct {
		Age  sql.NullString `dto:"age"`
		Name sql.NullString `dto:"name"`
	}{}
	if err := ConvertToDTO(userRequest, &userDto); err == nil {
		t.Fatal("should be types of fields are not identic error")
	}
}

func TestConvertorWhenStructureNotIdenticByTypesPointer(t *testing.T) {
	userRequest := struct {
		MyAge  *int64 `json:"age"`
		MyName string `json:"name"`
	}{}
	userDto := struct {
		Age  sql.NullInt64  `dto:"age"`
		Name sql.NullString `dto:"name"`
	}{}

	if err := ConvertToDTO(userRequest, &userDto); err != nil {
		t.Fatal(err.Error())
	}
}

func TestConvertorStringField(t *testing.T) {
	name := "Anatol"
	userRequest := struct {
		Age  *int64 `json:"age"`
		Name string `json:"name"`
	}{
		Age:  nil,
		Name: name,
	}
	userDto := struct {
		Age  sql.NullInt64  `dto:"age"`
		Name sql.NullString `dto:"name"`
	}{}

	if err := ConvertToDTO(userRequest, &userDto); err != nil {
		t.Fatal(err.Error())
	}

	if userDto.Name.String != userRequest.Name {
		t.Fatal("strings are not compared")
	}
}

func TestConvertorOtherTypesField(t *testing.T) {
	name := "Anatol"
	age := int64(1)
	userRequest := struct {
		Name string `json:"name"`
		Age  *int64 `json:"age"`
	}{
		Name: name,
		Age:  &age,
	}
	userDto := struct {
		Age  sql.NullInt64  `dto:"age"`
		Name sql.NullString `dto:"name"`
	}{}

	if err := ConvertToDTO(userRequest, &userDto); err != nil {
		t.Fatal(err.Error())
	}

	if userDto.Name.String != userRequest.Name {
		t.Fatal("strings are not compared")
	}
	if userDto.Age.Int64 != *userRequest.Age {
		t.Fatal("strings are not compared")
	}
}

func TestConvertorWithExistedBoolTypesField(t *testing.T) {
	name := "Anatol"
	userRequest := UserJsonRequest{
		Name: name,
	}
	userDto := UserDTO{}

	if err := ConvertToDTO(userRequest, &userDto); err != nil {
		t.Fatal(err.Error())
	}

	if userDto.Name.String != userRequest.Name {
		t.Fatal("strings are not compared")
	}
}

func TestConvertorWithNestedStructures(t *testing.T) {
	lvl1Data, lvl2Data := "level 1 Data", "level 2 Data"

	userRequest := struct {
		Level1 string `json:"level_1"`
		Level2 struct {
			Data *string `json:"data_level_2"`
		} `json:"my_nested"`
	}{
		Level1: lvl1Data,
		Level2: struct {
			Data *string `json:"data_level_2"`
		}{
			Data: &lvl2Data,
		},
	}

	userDto := struct {
		Level1 sql.NullString `dto:"level_1"`
		Level2 struct {
			Data sql.NullString `dto:"data_level_2"`
		} `dto:"my_nested"`
	}{}

	if err := ConvertToDTO(userRequest, &userDto); err != nil {
		t.Fatal(err.Error())
	}

	if userDto.Level1.String != userRequest.Level1 {
		t.Fatal("strings are not compared lvl 1")
	}
	if userDto.Level2.Data.String != *userRequest.Level2.Data {
		t.Fatal("strings are not compared lvl 2")
	}
}

func TestConvertorWithArraysStructures(t *testing.T) {
	lvl1Data, lvl2Data := "level 1 Data", "level 2 Data"

	userRequest := struct {
		Level1 string `json:"level_1"`
		Level2 []struct {
			Data *string `json:"data_level_2"`
		} `json:"my_array"`
		TestPtr *struct {
			Text string `json:"text"`
		} `json:"text_1"`
	}{
		Level1: lvl1Data,
		Level2: []struct {
			Data *string `json:"data_level_2"`
		}{
			{Data: &lvl2Data},
			{Data: nil},
		},
		TestPtr: &struct {
			Text string `json:"text"`
		}{Text: "text"},
	}

	userDto := struct {
		Level1 sql.NullString `dto:"level_1"`
		Level2 []struct {
			Data sql.NullString `dto:"data_level_2"`
		} `dto:"my_array"`
		TestPtr *struct {
			Text sql.NullString `dto:"text"`
		} `dto:"text_1"`
	}{}

	if err := ConvertToDTO(userRequest, &userDto); err != nil {
		t.Fatal(err.Error())
	}

	if userDto.Level1.String != userRequest.Level1 {
		t.Fatal("strings are not compared lvl 1 array")
	}
	if userDto.Level2[0].Data.String != *userRequest.Level2[0].Data {
		t.Fatal("strings are not compared lvl 2 array")
	}
	if userDto.Level2[1].Data.String != "" {
		t.Fatal("strings are not compared lvl 2 array")
	}
	if userDto.Level2[1].Data.Valid != false {
		t.Fatal("strings are not compared lvl 2 array")
	}
	if userDto.TestPtr.Text.String != userRequest.TestPtr.Text {
		t.Fatal("ptr are not compared")
	}
}

func TestConvertorTestArrays(t *testing.T) {
	jsonData := testStruct{
		Val1: "val1",
		Val2: &teststructneasted1{
			Val1: "val2",
			Val3: []string{"val2_1", "val2_2"},
		},
		ValArr: []teststructneasted1{
			{Val1: "varr1_1", Val3: []string{"varr1_1_1", "varr1_1_2", "varr1_1_3", "varr1_1_4"}},
			{Val1: "varr1_2", Val3: nil},
			{Val1: "varr1_3"},
		},
	}

	dtoData := testStructDto{}
	if err := ConvertToDTO(jsonData, &dtoData); err != nil {
		t.Fatal(err.Error())
	}
}

func TestArrays(t *testing.T) {
	jsonData := struct {
		Data []string `json:"data"`
	}{
		Data: []string{"one", "two"},
	}
	dtoData := struct {
		Data []sql.NullString `dto:"data"`
	}{}
	err := ConvertToDTO(jsonData, &dtoData)
	if err != nil {
		t.Error(err)
	}
}
