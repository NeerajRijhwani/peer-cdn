package torrent

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
)
type BencodeValue interface{}

func Decode(data io.Reader)(BencodeValue, error){
	char:=make([]byte,1)
	if _, err := data.Read(char); err != nil {
		return nil, err
	}
	switch char[0] {
	case 'd':
		return decodeDict(data)
	case 'l':
		return decodeList(data)
	case 'i':
		return decodeInt(data)
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return decodeString(data,char[0])
	default:
		return nil, fmt.Errorf("invalid bencode type: %c", char[0])
	}
}

func decodeInt(data io.Reader)(int64,error){
	char:=make([]byte,1)
	var number bytes.Buffer
	for{
		if _, err := data.Read(char); err != nil {
			return 0, err
		}
		if char[0] == 'e' {
			break
		}
		number.WriteByte(char[0])
	}
	return strconv.ParseInt(number.String(),10,64)
}

func decodeDict(data io.Reader)(map[string]BencodeValue,error){
	char:=make([]byte,1)
	dict:=make(map[string]BencodeValue)
	for{
		if _, err := data.Read(char); err != nil {
			return nil, err
		}
		if char[0] == 'e' {
			break
		}
		reader:=io.MultiReader(bytes.NewReader(char),data)
		keyValue,err:=Decode(reader)
		if err != nil {
			return nil, err
		}
		key, ok := keyValue.(string)
		if !ok {
			return nil, errors.New("dictionary key must be string")
		}
		value, err := Decode(data)
		if err != nil {
			return nil, err
		}
		dict[key] = value
	}
	return dict,nil
}

func decodeList(data io.Reader)([]BencodeValue,error){
	char:=make([]byte,1)
	var list []BencodeValue
	for{
		if _,err:=data.Read(char); err!=nil{
			return nil,err
		}
		if char[0]=='e'{
			break;
		}
		reader:=io.MultiReader(bytes.NewReader(char),data)
		value,err:=Decode(reader)
		if err!=nil{
			return nil,err
		}
		list = append(list, value)
	}

	return list,nil
}

func decodeString(data io.Reader,firstbyte byte)(string,error){
	char:=make([]byte,1)
	var lengthbuff bytes.Buffer
	lengthbuff.WriteByte(firstbyte)
	for{
		if _,err:=data.Read(char); err!=nil{
			return "",err
		}
		if(char[0]==':'){
			break
		}
		lengthbuff.WriteByte(char[0])
	}
	length, err := strconv.Atoi(lengthbuff.String())
	if err != nil {
		return "", err
	}
	str:=make([]byte,length)
	if _, err := io.ReadFull(data, str); err != nil {
		return "", err
	}

	return string(str), nil
}


func Encode(w io.Writer,val BencodeValue)error{
	switch v := val.(type) {
		case int:
			return encodeInt(w,v)
		case string:
			return encodeString(w,v)
		case []BencodeValue:
			return encodeList(w,v)
		case map[string]BencodeValue:
			return encodeDict(w,v)
		default:
			return fmt.Errorf("unsupported type: %T", v)
	}
}

func encodeInt(w io.Writer,val interface{}) error{
	var num int64
	switch v := val.(type) {
	case int:
		num = int64(v)
	case int64:
		num = v
	default:
		return errors.New("Not A Int")
	}

	_, err := fmt.Fprintf(w, "i%de", num)
	return err
}

func encodeString(w io.Writer,val string)error{
	_,err:=fmt.Fprintf(w,"%d:%s",len(val),val)
	return err
}

func encodeList(w io.Writer,val []BencodeValue)error{
	if _,err:=w.Write([]byte{'l'});err != nil {
		return err
	}
	for _, item := range val {
		if err:=Encode(w,item);err!=nil{
			return err
		}
	}
	_,err:=w.Write([]byte{'e'})
	return err
}


func encodeDict(w io.Writer,val map[string]BencodeValue) error{
	if _,err:=w.Write([]byte{'d'});err != nil {
		return err
	}

	keys := make([]string, 0, len(val))
	for k := range val {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	
	for _,key:=range keys{
		value:=val[key]
		if err := encodeString(w, key); err != nil {
			return err
		}
		if err:=Encode(w,value); err!=nil{
			return err
		}
	}

	_,err:=w.Write([]byte{'e'})
	return err
}