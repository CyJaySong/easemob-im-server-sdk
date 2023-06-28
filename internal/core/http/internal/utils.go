/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2021/8/16 3:47 下午
 * @Desc: TODO
 */

package internal

import (
	"encoding/json"
	"net/url"
	"strings"
)

const fileUploadingKey = "@file:"

func BuildParams(params interface{}) string {
	switch v := params.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case []interface{}:
		if len(v) > 0 {
			params = v[0]
		} else {
			params = nil
		}
	}

	m := make(map[string]interface{})

	if params != nil {
		if b, err := json.Marshal(params); err != nil {
			return String(params)
		} else if err = json.Unmarshal(b, &m); err != nil {
			return String(params)
		}
	} else {
		return ""
	}

	urlEncode := true

	if len(m) == 0 {
		return String(params)
	}

	for k, v := range m {
		if strings.Contains(k, fileUploadingKey) || strings.Contains(String(v), fileUploadingKey) {
			urlEncode = false
			break
		}
	}

	var (
		s   = ""
		str = ""
	)

	for k, v := range m {
		if len(str) > 0 {
			str += "&"
		}
		s = String(v)
		if urlEncode && len(s) > len(fileUploadingKey) && strings.Compare(s[0:len(fileUploadingKey)], fileUploadingKey) != 0 {
			s = url.QueryEscape(s)
		}
		str += k + "=" + s
	}

	return str
}