package registry

import (
	"fmt"
	"golang.org/x/sys/windows/registry"
)

const InternetSettings = "Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings"
const EnvironmentSettings = "Environment"
const notExist = "<NotExist>"
const RegKeyProxyEnable = "ProxyEnable"
const RegKeyProxyServer = "ProxyServer"
const RegKeyProxyOverride = "ProxyOverride"
const RegKeyHttpProxy = "HTTP_PROXY"

func SetGlobalProxy(port int, config *ProxyConfig) error {
	internetSettings, err := registry.OpenKey(registry.CURRENT_USER, InternetSettings, registry.ALL_ACCESS)
	if err != nil {
		return err
	}
	defer internetSettings.Close()

	val, _, err := internetSettings.GetIntegerValue(RegKeyProxyEnable)
	if err == nil {
		config.ProxyEnable = uint32(val)
	} else {
		config.ProxyEnable = 0
	}
	config.ProxyServer, _, err = internetSettings.GetStringValue(RegKeyProxyServer)
	if err != nil {
		config.ProxyServer = notExist
	}
	config.ProxyOverride, _, err = internetSettings.GetStringValue(RegKeyProxyOverride)
	if err != nil {
		config.ProxyOverride = notExist
	}

	internetSettings.SetDWordValue(RegKeyProxyEnable, 1)
	internetSettings.SetStringValue(RegKeyProxyServer, fmt.Sprintf("socks=127.0.0.1:%d", port))
	internetSettings.SetStringValue(RegKeyProxyOverride, "<local>")
	return nil
}

func CleanGlobalProxy(config *ProxyConfig) {
	internetSettings, err := registry.OpenKey(registry.CURRENT_USER, InternetSettings, registry.ALL_ACCESS)
	if err == nil {
		defer internetSettings.Close()
		internetSettings.SetDWordValue(RegKeyProxyEnable, config.ProxyEnable)
		if config.ProxyServer != notExist {
			internetSettings.SetStringValue(RegKeyProxyServer, config.ProxyServer)
		} else {
			internetSettings.DeleteValue(RegKeyProxyServer)
		}
		if config.ProxyOverride != notExist {
			internetSettings.SetStringValue(RegKeyProxyOverride, config.ProxyOverride)
		} else {
			internetSettings.DeleteValue(RegKeyProxyOverride)
		}
	}
}

func SetHttpProxyEnvironmentVariable(port int, config *ProxyConfig) error {
	internetSettings, err := registry.OpenKey(registry.CURRENT_USER, EnvironmentSettings, registry.ALL_ACCESS)
	if err != nil {
		return err
	}
	defer internetSettings.Close()

	config.HttpProxyVar, _, err = internetSettings.GetStringValue(RegKeyHttpProxy)
	if err != nil {
		config.HttpProxyVar = notExist
	}

	internetSettings.SetStringValue(RegKeyHttpProxy, fmt.Sprintf("socks://127.0.0.1:%d", port))
	refreshEnvironmentVariable()
	return nil
}

func CleanHttpProxyEnvironmentVariable(config *ProxyConfig) {
	internetSettings, err := registry.OpenKey(registry.CURRENT_USER, EnvironmentSettings, registry.ALL_ACCESS)
	if err == nil {
		defer internetSettings.Close()
		if config.HttpProxyVar != notExist {
			internetSettings.SetStringValue(RegKeyHttpProxy, config.HttpProxyVar)
		} else {
			internetSettings.DeleteValue(RegKeyHttpProxy)
		}
		refreshEnvironmentVariable()
	}
}

func refreshEnvironmentVariable() {
	syscall.NewLazyDLL("user32.dll").NewProc("SendMessageW").Call(
		HWND_BROADCAST, WM_SETTINGCHANGE, 0, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("ENVIRONMENT"))))
}
