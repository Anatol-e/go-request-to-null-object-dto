package convertor

import (
	"database/sql"
	"errors"
	"fmt"
	errs "github.com/pkg/errors"
	"reflect"
	"strings"
)

const (
	dtoTag     = "dto"
	jsonTag    = "json"
	FieldIndex = 0
	ValidIndex = 1
)

type nullPrimitive struct {
	primitiveType string
	isPointer     bool
	valueIndex    int
}

type actionSetter struct {
	typeDto reflect.Type
	action  func(dto reflect.Value, jsonField reflect.Value)
}

var mapOfNullTypes = map[string]actionSetter{
	"string":  {typeDto: reflect.TypeOf(sql.NullString{}), action: func(d reflect.Value, j reflect.Value) { d.SetString(j.String()) }},
	"int64":   {typeDto: reflect.TypeOf(sql.NullInt64{}), action: func(d reflect.Value, j reflect.Value) { d.SetInt(j.Int()) }},
	"int":     {typeDto: reflect.TypeOf(sql.NullInt32{}), action: func(d reflect.Value, j reflect.Value) { d.SetInt(j.Int()) }},
	"int32":   {typeDto: reflect.TypeOf(sql.NullInt32{}), action: func(d reflect.Value, j reflect.Value) { d.SetInt(j.Int()) }},
	"float64": {typeDto: reflect.TypeOf(sql.NullFloat64{}), action: func(d reflect.Value, j reflect.Value) { d.SetFloat(j.Float()) }},
	"float32": {typeDto: reflect.TypeOf(sql.NullFloat64{}), action: func(d reflect.Value, j reflect.Value) { d.SetFloat(j.Float()) }},
	"bool":    {typeDto: reflect.TypeOf(sql.NullBool{}), action: func(d reflect.Value, j reflect.Value) { d.SetBool(j.Bool()) }},
}

type reflectStuct struct {
	t reflect.Type
	v reflect.Value
}

func (tv *reflectStuct) size() int {
	return tv.t.NumField()
}

func ConvertToDTO(jsonObject interface{}, dtoObject interface{}) error {
	if jsonObject == nil || dtoObject == nil {
		return errors.New("nil objects can't be processed")
	}

	tvJson := reflectStuct{
		t: reflect.TypeOf(jsonObject),
		v: reflect.ValueOf(jsonObject),
	}
	tvDto := reflectStuct{
		t: reflect.TypeOf(dtoObject).Elem(),
		v: reflect.ValueOf(dtoObject).Elem(),
	}
	_, err := convert(&tvJson, &tvDto)
	return err
}

func convert(tvJson *reflectStuct, tvDto *reflectStuct) (*reflectStuct, error) {
	mapOfStructures, err := validateObjectComparing(tvJson, tvDto)
	if err != nil {
		return nil, err
	}
	return setData(mapOfStructures, tvJson, tvDto), nil
}

func setData(mapOfStructures map[nullPrimitive]int, tvJson *reflectStuct, tvDto *reflectStuct) *reflectStuct {
	for jsonReflObject, dtoReflObject := range mapOfStructures {
		dtoAction := mapOfNullTypes[jsonReflObject.primitiveType]
		jsonFieldValue := tvJson.v.Field(jsonReflObject.valueIndex)
		isNil := false
		if jsonReflObject.isPointer {
			isNil = jsonFieldValue.IsNil()
			jsonFieldValue = jsonFieldValue.Elem()
		}
		if !isNil {
			dtoAction.action(tvDto.v.Field(dtoReflObject).Field(FieldIndex), jsonFieldValue)
		}
		tvDto.v.Field(dtoReflObject).Field(ValidIndex).SetBool(!jsonReflObject.isPointer)
	}
	return tvDto
}

func validateObjectComparing(tvJson *reflectStuct, tvDto *reflectStuct) (map[nullPrimitive]int, error) {
	if ok := setPrimitiveValueWithoutProcessing(tvJson, tvDto); ok {
		return nil, nil
	}
	if tvJson.size() != tvDto.size() {
		return nil, errors.New("number of fields are not identic")
	}

	jsonDtoTypesMap := make(map[nullPrimitive]int, tvDto.size())
	for i := 0; i < tvDto.size(); i++ {
		tvJsonTagField := tvJson.t.Field(i)
		tvDtoFieldIndex, err := getDtoFieldIndexByTag(tvDto.t, tvJsonTagField.Tag.Get(jsonTag))
		if err != nil {
			return nil, err
		}

		if tvJson.v.Field(i).Kind() == reflect.Slice {
			if err = processSlice(tvJson, tvDto, i); err != nil {
				return nil, err
			}
			continue
		}
		if tvDto.v.Field(tvDtoFieldIndex).Type().Kind() == reflect.Ptr {
			if !tvJson.v.Field(i).IsNil() {
				dtoValue := reflect.New(tvDto.v.Field(i).Type().Elem())
				tvDto.v.Field(i).Set(dtoValue)
				if err = processPtr(tvJson, tvDto, i); err != nil {
					return nil, err
				}
			}
			continue
		}
		if tvJson.v.Field(i).Type().Kind() == reflect.Struct {
			if err = processStruct(tvJson, tvDto, i); err != nil {
				return nil, err
			}
			continue
		}

		jsonType := tvJsonTagField.Type.String()
		isPointer := strings.Contains(jsonType, "*")
		if isPointer {
			jsonType = jsonType[1:]
		}
		jsonDtoTypesMap[nullPrimitive{
			primitiveType: jsonType,
			isPointer:     isPointer,
			valueIndex:    i,
		}] = tvDtoFieldIndex

		if mapOfNullTypes[jsonType].typeDto != tvDto.t.Field(tvDtoFieldIndex).Type {
			return jsonDtoTypesMap, errors.New("types of fields are not identic with DTO object")
		}
	}
	return jsonDtoTypesMap, nil
}

func setPrimitiveValueWithoutProcessing(tvJson *reflectStuct, tvDto *reflectStuct) bool {
	pr := mapOfNullTypes[tvJson.v.Type().String()]
	if pr.action == nil {
		return false
	}
	pr.action(tvDto.v.Field(FieldIndex), tvJson.v)
	tvDto.v.Field(ValidIndex).SetBool(true)
	return true
}

func processPtr(tvJson *reflectStuct, tvDto *reflectStuct, i int) error {
	tvDtoNested, err := convert(
		&reflectStuct{
			t: tvJson.t.Field(i).Type.Elem(),
			v: tvJson.v.Field(i).Elem(),
		},
		&reflectStuct{
			t: tvDto.t.Field(i).Type.Elem(),
			v: tvDto.v.Field(i).Elem(),
		})
	if err != nil {
		return errs.Wrap(err, "nested level error")
	}
	tvDto = tvDtoNested
	return nil
}

func processStruct(tvJson *reflectStuct, tvDto *reflectStuct, i int) error {
	tvDtoNested, err := convert(
		&reflectStuct{
			t: tvJson.t.Field(i).Type,
			v: tvJson.v.Field(i),
		},
		&reflectStuct{
			t: tvDto.t.Field(i).Type,
			v: tvDto.v.Field(i),
		})
	if err != nil {
		return errs.Wrap(err, "nested level error")
	}
	tvDto = tvDtoNested
	return nil
}

func processSlice(tvJson *reflectStuct, tvDto *reflectStuct, i int) error {
	sliceSize := tvJson.v.Field(i).Len()
	reflectSlice := reflect.MakeSlice(reflect.SliceOf(tvDto.v.Field(i).Type().Elem()), 0, 0)
	for sliceIndex := 0; sliceIndex < sliceSize; sliceIndex++ {
		dtoValue := reflect.New(tvDto.v.Field(i).Type().Elem())
		tvDtoNested, err := convert(
			&reflectStuct{
				t: tvJson.v.Field(i).Index(sliceIndex).Type(),
				v: tvJson.v.Field(i).Index(sliceIndex),
			},
			&reflectStuct{
				t: dtoValue.Type().Elem(),
				v: dtoValue.Elem(),
			})
		if err != nil {
			return errs.Wrap(err, "nested level error")
		}
		reflectSlice = reflect.Append(reflectSlice, reflect.ValueOf(tvDtoNested.v.Interface()))
	}
	tvDto.v.Field(i).Set(reflectSlice)
	return nil
}

func getDtoFieldIndexByTag(field reflect.Type, tagValue string) (int, error) {
	for i := 0; i < field.NumField(); i++ {
		if field.Field(i).Tag.Get(dtoTag) == tagValue {
			return i, nil
		}
	}
	return 0, errors.New(fmt.Sprintf("tag [%s] is not exists in dto object. Should be the same numbers of fields and tags", tagValue))
}
