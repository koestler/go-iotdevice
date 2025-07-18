include $(TOPDIR)/rules.mk

PKG_NAME:=go-iotdevice
PKG_VERSION:=1.0.0
PKG_RELEASE:=1

PKG_SOURCE_PROTO:=git
PKG_SOURCE_URL:=https://github.com/koestler/go-iotdevice.git
PKG_SOURCE_VERSION:=HEAD
PKG_SOURCE:=$(PKG_NAME)-$(PKG_VERSION).tar.gz
PKG_SOURCE_SUBDIR:=$(PKG_NAME)-$(PKG_VERSION)

PKG_BUILD_DEPENDS:=golang/host
PKG_BUILD_PARALLEL:=1
PKG_USE_MIPS16:=0

GO_PKG:=github.com/koestler/go-iotdevice

include $(INCLUDE_DIR)/package.mk
include ../../lang/golang/golang-package.mk

define Package/go-iotdevice
  SECTION:=utils
  CATEGORY:=Utilities
  TITLE:=go-iotdevice
  URL:=https://github.com/koestler/go-iotdevice
  DEPENDS:=$(GO_ARCH_DEPENDS)
endef

define Package/go-iotdevice/description
  A Go-based tool to read values from various IoT devices (Victron, Shelly, Modbus, etc.) and publish
  them to MQTT servers. Provides a REST and websocket API, and an internal web frontend. Designed for
  off-grid solar, battery, and generator monitoring/control. No historical data storage; integrates
  with InfluxDB and Grafana for dashboards.
endef

define Build/Compile
	$(call GoPackage/Build/Compile)
endef

define Package/go-iotdevice/install
	$(INSTALL_DIR) $(1)/usr/bin
	$(INSTALL_BIN) $(GO_PKG_BUILD_BIN_DIR)/go-iotdevice $(1)/usr/bin/
	$(INSTALL_DIR) $(1)/etc/init.d
	$(INSTALL_BIN) ./files/go-iotdevice.init $(1)/etc/init.d/go-iotdevice
endef

$(eval $(call GoBinPackage,go-iotdevice))
$(eval $(call BuildPackage,go-iotdevice))
