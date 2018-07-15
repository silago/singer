package main

import (
	"log"
	"os"
    "encoding/json"
    "fmt"    
    "sort"
    "strings"
    "crypto/md5"
    "encoding/hex"
	"net/http"
)

func ENV(name string) string {
    result:=""
    if s, ok := os.LookupEnv(name); ok {
        result = s
    } else {
        log.Fatal("Could not get env var " +  name)
    }
    return result
}


type signer struct {
}

func (s signer) prepareArray(controller string, action string, input_params map[string]interface{}) ( map[string]interface{}, error ) {
    result_params:=input_params
    result_params["controller"]=controller
    result_params["action"]=action
    return result_params, nil
}

func mergeMap(a map[string]string, b map[string]string) map[string]string {
    for k, v:= range b {
        a[k] = v
    }
    return a
}

func  (s signer) calcArraySignMap(params map[string]interface{}) map[string]string {
    result:=make(map[string]string)
    for _, key:= range s.getSortedKeys(params) {
        val:=params[key]
        switch concreteVal := val.(type) {
            case map[string]interface{}:
                result = mergeMap(result, s.calcArraySignMap(val.(map[string]interface{})))
            case []interface{}:
                result = mergeMap(result, s.calcArraySignArray(val.([]interface{})))
            default:
                result[key] = fmt.Sprint(concreteVal)
        }
    }
    return result
}

func  (s signer) calcArraySignArray(anArray []interface{}) map[string]string {
    result:=make(map[string]string)
    for _, val := range anArray {
        switch concreteVal := val.(type) {
        case map[string]interface{}:
            result = mergeMap(result, s.calcArraySignMap(val.(map[string]interface{})))
        case []interface{}:
            result = mergeMap(result, s.calcArraySignArray(val.([]interface{})))
        default:
            fmt.Println(">>>!!",concreteVal)
        }
    }
    return result
}


func  (s signer) getSortedKeys(params map[string]interface{}) []string { 
    var keys []string
    for k := range params {
        keys = append(keys, k)
    }
    sort.Strings(keys)
    return keys
}

func hashMd5(input string) string {
    hasher := md5.New()
    hasher.Write([]byte(input))
    return hex.EncodeToString(hasher.Sum(nil))
}

func  (s signer) getControllerActionNames(req *http.Request) (string, string) {
    ps := strings.SplitN(req.URL.Path, "/", 3)
    controller:=ps[1]
    action:= ps[2]
    return controller, action
}

func  (s signer) Execute () ( string ) {
    var data map[string]interface{}        
    args := os.Args[1:]
    controller:=args[0]
    action:=args[1]
    request_body:=args[2]
    if err := json.Unmarshal([]byte(request_body), &data); err != nil {
        panic(err)
    }

    params, _ := s.prepareArray(controller,action,data)
    sign:= s.calcArraySignMap(params)
    names := make([]string, 0, len(sign))
    for name := range sign {
        names = append(names, name)
    }
    sort.Strings(names)
    var result []string

    for _, name := range names {
        result = append(result, name+"="+sign[name])
    }
    return  hashMd5(strings.Join(result,"&"))
}

func (s signer) InitSigner() func() ( string ) {
    return s.Execute
}

func main() {
    _signer:= signer{} 
    sign:=_signer.Execute()
    fmt.Println("",sign)
}
