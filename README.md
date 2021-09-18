# go-request-to-null-object-dto

## Object from JSON to DTO (NullObject's)

This library helps to simply fill Nullable structure from json object. It can be used for example if we want to use
structure without pointers after request in golang.

## How to use

```ConvertToDTO(jsonObject interface{}, dtoObject interface{}) error```
can convert any json request to DTO without pointers.

* Create JSON object

```
type SomeRequest struct {
    Data  string `json:"data"`  
    Value int64  `json:"value"`  
}
```

* Create DTO object:
    * Create identically (recommended used same fields order)
    * Mark fields with the tag "dto" with the same values as from json structure

```
type SomeDTO struct {
    MyData  string `dto:"data"`    
    MyValue string `dto:"value"`    
}
```

* Use it by using converting method:

```
convertor.ConvertToDTO(jsonObject, &dtoObject)
fmt.Println(dtoObject.Data.String)
```

## Nested structures using

Json structure should be the same.

```
type SomeDTO struct {
    MyData OtherStruct `dto:"data"`    
}
```

Also can be nullable

```
type SomeDTO struct {
    MyData *OtherStruct `dto:"data"`    
}
```

## Arrays using

```
type SomeDTO struct {
    MyData []sql.NullString `dto:"data"`    
}
```

Also with structures

```
type SomeDTO struct {
    MyData []OtherStruct `dto:"data"`    
}
```
