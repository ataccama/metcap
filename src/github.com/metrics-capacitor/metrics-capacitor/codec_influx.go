package metcap

//
// import (
// 	"fmt"
// 	"regexp"
// 	"strconv"
// 	"strings"
// 	"time"
// )
//
// // helper function to parse timestamp into time.Time
// func parseTimestamp(d map[string]string) time.Time {
// 	var (
// 		t_now       time.Time
// 		t_str       string
// 		t_byte      []byte
// 		t_len       int
// 		t_unix_sec  int64
// 		t_unix_nsec int64
// 		err         error
// 	)
//
// 	t_now = time.Now()
// 	t_str = d["timestamp"]
// 	t_byte = []byte(t_str)
// 	t_len = len(t_byte)
//
// 	switch {
// 	// time not specified
// 	case t_len == 0:
// 		return t_now
// 	// time is in Unix timestamp
// 	case t_len <= 11:
// 		t_int, err := strconv.ParseInt(t_str, 10, 64)
// 		if err != nil {
// 			return t_now
// 		}
// 		return time.Unix(t_int, 0)
// 	// time is in Unix timestamp with second fractions
// 	case t_len > 11:
// 		t_unix_sec, err = strconv.ParseInt(string(t_byte[:11]), 10, 64)
// 		if err != nil {
// 			return t_now
// 		} else {
// 			t_unix_nsec, err = strconv.ParseInt(string(t_byte[11:t_len])+strings.Repeat("0", len(t_byte[11:t_len])), 10, 64)
// 			if err != nil {
// 				return t_now
// 			}
// 		}
// 		return time.Unix(t_unix_sec, t_unix_nsec*int64(time.Millisecond))
// 	default:
// 		return t_now
// 	}
// }
//
// // helper function to parse value as float64
// func parseValue(d map[string]string) (float64, error) {
// 	var (
// 		value float64
// 		err   error
// 	)
// 	if value, err = strconv.ParseFloat(d["value"], 64); err != nil {
// 		return float64(0), &ParserError{"Failed to parse value", d["value"]}
// 	}
// 	return value, nil
// }
//
// // helper function to parse metric name and fields
// func parseFields(d map[string]string, mut *[]string) (string, map[string]string, error) {
// 	name := []string{}
// 	fields := make(map[string]string)
// 	mut_rule_match := false
//
// 	// check if we have graphite path
// 	if _, ok := d["path"]; ok {
// 		// iterate through mutator rules
// 		for _, lineRule := range *mut {
// 			mut_rule := strings.Split(lineRule, "|||")
// 			mut_re, err := regexp.Compile(mut_rule[0])
// 			if err != nil {
// 				continue
// 			}
// 			// try to match metric path with a mutator rule
// 			if mut_re.Match([]byte(d["path"])) {
// 				mut_rule_match = true
// 				field_values := strings.Split(d["path"], ".")
// 				field_names := strings.Split(mut_rule[1], ".")
//
// 				// iterate thru fields
// 			FIELD_PARSER:
// 				for i, field := range field_values {
// 					switch {
// 					case regexp.MustCompile(`^[0-9]+$`).Match([]byte(field_names[i])):
// 						// numeric rule -> name
// 						name = append(name, field)
// 					case regexp.MustCompile(`^[a-zA-Z0-9_]+\+$`).Match([]byte(field_names[i])):
// 						// string rule with catch-all flag -> catch-all field
// 						f := strings.TrimRight(field_names[i], "+")
// 						fields[f] = strings.Join(field_values[i:], ":")
// 						break FIELD_PARSER
// 					case regexp.MustCompile(`^[a-zA-Z0-9_]+$`).Match([]byte(field_names[i])):
// 						// string rule -> field
// 						fields[field_names[i]] = field
// 					case regexp.MustCompile(`^\+$`).Match([]byte(field_names[i])):
// 						// catch-all flag -> fill name
// 						name = append(name, strings.Join(field_values[i:], ":"))
// 						break FIELD_PARSER
// 					case regexp.MustCompile(`^-$`).Match([]byte(field_names[i])):
// 						// no-catch flag -> skip
// 						continue FIELD_PARSER
// 					}
// 				}
// 				break
// 			}
// 		}
//
// 		if !mut_rule_match {
// 			name = append(name, strings.Join(strings.Split(d["path"], "."), ":"))
// 		}
// 		// not Graphite? then it must be only Influx (for now :))
// 	} else {
// 		name = append(name, d["name"])
// 		// iterate thru fields
// 		for _, field := range strings.Split(d["fields"], ",") {
// 			kv := strings.Split(field, "=")
// 			if kv[0] != "" {
// 				fields[kv[0]] = kv[1]
// 			}
// 		}
// 	}
// 	if len(name) == 0 {
// 		return "", make(map[string]string), &ParserError{"Failed to parse metric name", name}
// 	}
// 	return strings.Join(name, ":"), fields, nil
// }
//
// type ParserError struct {
// 	msg string
// 	src interface{}
// }
//
// func (e *ParserError) Error() string {
// 	return fmt.Sprintf("%s - %v", e.msg, e.src)
// }
