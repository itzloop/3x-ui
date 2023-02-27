package service

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"reflect"
	"strconv"
	"strings"
	"time"
	"x-ui/database"
	"x-ui/database/model"
	"x-ui/logger"
	"x-ui/util/common"
	"x-ui/util/random"
	"x-ui/util/reflect_util"
	"x-ui/web/entity"
	"x-ui/web/global"
	"x-ui/web/session"
)

//go:embed config.json
var xrayTemplateConfig string

var defaultValueMap = map[string]string{
	"xrayTemplateConfig": xrayTemplateConfig,
	"webListen":          "",
	"webPort":            "2053",
	"webCertFile":        "",
	"webKeyFile":         "",
	"secret":             random.Seq(32),
	"webBasePath":        "/",
	"timeLocation":       "Asia/Tehran",
	"tgBotEnable":        "false",
	"tgBotToken":         "",
	"tgBotChatId":        "0",
	"tgRunTime":          "",
}

type SettingService struct {
}

func (s *SettingService) GetAllSetting() (*entity.AllSetting, error) {
	db := database.GetDB()
	settings := make([]*model.Setting, 0)
	err := db.Model(model.Setting{}).Find(&settings).Error
	if err != nil {
		return nil, err
	}
	allSetting := &entity.AllSetting{}
	t := reflect.TypeOf(allSetting).Elem()
	v := reflect.ValueOf(allSetting).Elem()
	fields := reflect_util.GetFields(t)

	setSetting := func(key, value string) (err error) {
		defer func() {
			panicErr := recover()
			if panicErr != nil {
				err = errors.New(fmt.Sprint(panicErr))
			}
		}()

		var found bool
		var field reflect.StructField
		for _, f := range fields {
			if f.Tag.Get("json") == key {
				field = f
				found = true
				break
			}
		}

		if !found {
			// Some settings are automatically generated, no need to return to the front end to modify the user
			return nil
		}

		fieldV := v.FieldByName(field.Name)
		switch t := fieldV.Interface().(type) {
		case int:
			n, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			fieldV.SetInt(n)
		case string:
			fieldV.SetString(value)
		case bool:
			fieldV.SetBool(value == "true")
		default:
			return common.NewErrorf("unknown field %v type %v", key, t)
		}
		return
	}

	keyMap := map[string]bool{}
	for _, setting := range settings {
		err := setSetting(setting.Key, setting.Value)
		if err != nil {
			return nil, err
		}
		keyMap[setting.Key] = true
	}

	for key, value := range defaultValueMap {
		if keyMap[key] {
			continue
		}
		err := setSetting(key, value)
		if err != nil {
			return nil, err
		}
	}

	return allSetting, nil
}

func (s *SettingService) ResetSettings() error {
	db := database.GetDB()
	return db.Where("1 = 1").Delete(model.Setting{}).Error
}

func (s *SettingService) getSetting(key string) (*model.Setting, error) {
	db := database.GetDB()
	setting := &model.Setting{}
	err := db.Model(model.Setting{}).Where("key = ?", key).First(setting).Error
	if err != nil {
		return nil, err
	}
	return setting, nil
}

func (s *SettingService) saveSetting(key string, value string) error {
	setting, err := s.getSetting(key)
	db := database.GetDB()
	if database.IsNotFound(err) {
		return db.Create(&model.Setting{
			Key:   key,
			Value: value,
		}).Error
	} else if err != nil {
		return err
	}
	setting.Key = key
	setting.Value = value
	return db.Save(setting).Error
}

func (s *SettingService) getString(key string) (string, error) {
	setting, err := s.getSetting(key)
	if database.IsNotFound(err) {
		value, ok := defaultValueMap[key]
		if !ok {
			return "", common.NewErrorf("key <%v> not in defaultValueMap", key)
		}
		return value, nil
	} else if err != nil {
		return "", err
	}
	return setting.Value, nil
}

func (s *SettingService) setString(key string, value string) error {
	return s.saveSetting(key, value)
}

func (s *SettingService) getBool(key string) (bool, error) {
	str, err := s.getString(key)
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(str)
}

func (s *SettingService) setBool(key string, value bool) error {
	return s.setString(key, strconv.FormatBool(value))
}

func (s *SettingService) getInt(key string) (int, error) {
	str, err := s.getString(key)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(str)
}

func (s *SettingService) setInt(key string, value int) error {
	return s.setString(key, strconv.Itoa(value))
}

func (s *SettingService) GetXrayConfigTemplate() (string, error) {
	return s.getString("xrayTemplateConfig")
}

func (s *SettingService) GetListen() (string, error) {
	return s.getString("webListen")
}

func (s *SettingService) GetTgBotToken() (string, error) {
	return s.getString("tgBotToken")
}

func (s *SettingService) SetTgBotToken(token string) error {
	return s.setString("tgBotToken", token)
}

func (s *SettingService) GetTgBotChatId() (int, error) {
	return s.getInt("tgBotChatId")
}

func (s *SettingService) SetTgBotChatId(chatId int) error {
	return s.setInt("tgBotChatId", chatId)
}

func (s *SettingService) SetTgbotenabled(value bool) error {
	return s.setBool("tgBotEnable", value)
}

func (s *SettingService) GetTgbotenabled() (bool, error) {
	return s.getBool("tgBotEnable")
}

func (s *SettingService) SetTgbotRuntime(time string) error {
	return s.setString("tgRunTime", time)
}

func (s *SettingService) GetTgbotRuntime() (string, error) {
	return s.getString("tgRunTime")
}

func (s *SettingService) GetPort() (int, error) {
	return s.getInt("webPort")
}

func (s *SettingService) SetPort(port int) error {
	return s.setInt("webPort", port)
}

func (s *SettingService) GetCertFile() (string, error) {
	return s.getString("webCertFile")
}

func (s *SettingService) GetKeyFile() (string, error) {
	return s.getString("webKeyFile")
}

func (s *SettingService) GetSecret() ([]byte, error) {
	secret, err := s.getString("secret")
	if secret == defaultValueMap["secret"] {
		err := s.saveSetting("secret", secret)
		if err != nil {
			logger.Warning("save secret failed:", err)
		}
	}
	return []byte(secret), err
}

func (s *SettingService) GetBasePath() (string, error) {
	basePath, err := s.getString("webBasePath")
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(basePath, "/") {
		basePath = "/" + basePath
	}
	if !strings.HasSuffix(basePath, "/") {
		basePath += "/"
	}
	return basePath, nil
}

func (s *SettingService) GetTimeLocation() (*time.Location, error) {
	l, err := s.getString("timeLocation")
	if err != nil {
		return nil, err
	}
	location, err := time.LoadLocation(l)
	if err != nil {
		defaultLocation := defaultValueMap["timeLocation"]
		logger.Errorf("location <%v> not exist, using default location: %v", l, defaultLocation)
		return time.LoadLocation(defaultLocation)
	}
	return location, nil
}

func (s *SettingService) UpdateAllSetting(c *gin.Context, allSetting *entity.AllSetting) error {
	if err := allSetting.CheckValid(); err != nil {
		return err
	}

	v := reflect.ValueOf(allSetting).Elem()
	t := reflect.TypeOf(allSetting).Elem()
	fields := reflect_util.GetFields(t)
	errs := make([]error, 0)
	for _, field := range fields {
		key := field.Tag.Get("json")
		fieldV := v.FieldByName(field.Name)
		value := fmt.Sprint(fieldV.Interface())

		if key == "xrayTemplateConfig" {
			if err := s.handleInboundsFromConfig(c, value); err != nil {
				return err
			}
		}
		err := s.saveSetting(key, value)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return common.Combine(errs...)
}

func (s *SettingService) handleInboundsFromConfig(c *gin.Context, value string) error {
	// update inbounds :)
	// get inbounds from value
	configMap := map[string]interface{}{}
	if err := json.Unmarshal([]byte(value), &configMap); err != nil {
		return err
	}

	if _, ok := configMap["inbounds"]; !ok {
		return nil
	}

	if _, ok := configMap["inbounds"].([]interface{}); !ok {
		return nil
	}

	var inbounds []*model.Inbound
	for _, inboundInr := range configMap["inbounds"].([]interface{}) {
		inbound, ok := inboundInr.(map[string]interface{})
		if !ok {
			continue
		}

		user := session.GetLoginUser(c)
		var inboundStruct model.Inbound
		inboundStruct.UserId = user.Id
		up, _ := inbound["up"].(float64)
		inboundStruct.Up = int64(up)
		down, _ := inbound["down"].(float64)
		inboundStruct.Down = int64(down)
		total, _ := inbound["total"].(float64)
		inboundStruct.Total = int64(total)
		inboundStruct.Remark, _ = inbound["remark"].(string)
		inboundStruct.Enable, _ = inbound["enable"].(bool)
		expiryTime, _ := inbound["expiryTime"].(float64)
		inboundStruct.ExpiryTime = int64(expiryTime)
		inboundStruct.Listen, _ = inbound["listen"].(string)
		port, _ := inbound["port"].(float64)
		inboundStruct.Port = int(port)
		pr, _ := inbound["protocol"].(string)
		inboundStruct.Protocol = model.Protocol(pr)

		j, _ := json.Marshal(inbound["settings"])
		if string(j) == "null" {
			continue
		}
		inboundStruct.Settings = string(j)

		j, _ = json.Marshal(inbound["streamSettings"])
		if string(j) == "null" {
			continue
		}
		inboundStruct.StreamSettings = string(j)
		inboundStruct.Tag, _ = inbound["tag"].(string)

		j, _ = json.Marshal(inbound["sniffing"])
		if string(j) == "null" {
			continue
		}
		inboundStruct.Sniffing = string(j)

		if inboundStruct.Port == 0 && strings.TrimSpace(inboundStruct.Remark) == "" {
			inboundStruct.Remark = strings.ToLower(fmt.Sprintf("%s-fallback", inboundStruct.Tag))
		} else if strings.TrimSpace(inboundStruct.Remark) == "" {
			inboundStruct.Remark = strings.ToLower(inboundStruct.Tag)
		}

		inbounds = append(inbounds, &inboundStruct)
	}

	if err := global.GetInbounds().AddInbounds(inbounds); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "port") {
			return err
		}
	}

	return nil
}
