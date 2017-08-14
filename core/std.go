package core

import (
	"time"

	"github.com/crazyprograms/sud/callpull"
	"github.com/crazyprograms/sud/corebase"
)

func stdConfigurationList(cr *Core, Name string, Param map[string]interface{}, timeOutWait time.Duration, Access corebase.IAccess) (callpull.Result, error) {
	configurations := cr.GetConfiguration()
	listConf := make(map[string]interface{})
	for name := range configurations {
		var err error
		var confInfo map[string]interface{}
		if confInfo, err = configurations[name].Save(); err != nil {
			return callpull.Result{Error: err, Result: nil}, nil
		}
		listConf[name] = confInfo
	}
	return callpull.Result{Error: nil, Result: listConf}, nil
}

func InitStdModule(c *Core) bool {
	conf := NewConfiguration([]string{"std"})
	conf.AddCall(CallInfo{ConfigurationName: "std", Name: "std.Configuration.Get", PullName: "std", AccessCall: "Configuration", AccessListen: "", Title: "Список конфигураций"})
	if AddStdCall("std.Configuration.Get", stdConfigurationList) != nil {
		return false
	}
	return c.AddBaseConfiguration("std", conf)
}