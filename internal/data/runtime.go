package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

)

type Runtime int32

var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

func (r Runtime) MarshalJSON() ([]byte, error){
  // generate a runtime string with appropriate format required
  jsonVal := fmt.Sprintf("%d mins", r)

  quotedJsonVal := strconv.Quote(jsonVal)

  return []byte(quotedJsonVal), nil
}


func (r *Runtime) UnmarshalJSON(jsonVal []byte) error{
  // we expect incoming request with the value "<value> mins" and the first thing we need to do is remove the quotes 
  unquoteJsonVal , err := strconv.Unquote(string(jsonVal))
  if err != nil {
    return ErrInvalidRuntimeFormat
  }
   // split the string with " "
   parts := strings.Split(unquoteJsonVal, " ")

   // check if the parths length is 2 for value and mins
   // if the length is not 2 return error invalid format
   if len(parts) != 2 || parts[1] != "mins" {
     return ErrInvalidRuntimeFormat
   } 

   runtime , err := strconv.ParseInt(parts[0],10,32)
   if err != nil {
     return ErrInvalidRuntimeFormat
   }


   *r = Runtime(runtime)
   return nil 

}


