# Changelog

## [1.0.12](https://github.com/SpechtLabs/rpi_exporter/compare/v1.0.11...v1.0.12) (2026-10-10)


### Bug Fixes

* **deps:** update module github.com/prometheus/client_golang to v1.25.0 ([#52](https://github.com/SpechtLabs/rpi_exporter/issues/52)) ([0211ee4](https://github.com/SpechtLabs/rpi_exporter/commit/0211ee4755cdadb428866237a12bd7f7e1c360f3))

## [1.0.11](https://github.com/SpechtLabs/rpi_exporter/compare/v1.0.10...v1.0.11) (2026-10-05)


### Bug Fixes

* answer an unknown or disabled collect[] value with the valid ones, as plain text ([9b23910](https://github.com/SpechtLabs/rpi_exporter/commit/9b23910deef446360327f7b20dd61d2b53765169))
* declare the module path github.com/spechtlabs/rpi_exporter ([9b23910](https://github.com/SpechtLabs/rpi_exporter/commit/9b23910deef446360327f7b20dd61d2b53765169))
* **deps:** update all Go dependencies ([9b23910](https://github.com/SpechtLabs/rpi_exporter/commit/9b23910deef446360327f7b20dd61d2b53765169))
* **deps:** update go modules ([#44](https://github.com/SpechtLabs/rpi_exporter/issues/44)) ([ab745e1](https://github.com/SpechtLabs/rpi_exporter/commit/ab745e1cb200605a23c32882a57ff693722c7391))
* **deps:** update module github.com/prometheus/client_golang to v1.24.0 ([#36](https://github.com/SpechtLabs/rpi_exporter/issues/36)) ([7925b74](https://github.com/SpechtLabs/rpi_exporter/commit/7925b748e06a503414fed9f0809419ee6944b2c6))
* **deps:** update module github.com/prometheus/client_golang to v1.24.1 ([#38](https://github.com/SpechtLabs/rpi_exporter/issues/38)) ([80046a0](https://github.com/SpechtLabs/rpi_exporter/commit/80046a0b7c5d35979de7673f987680bc2b84cc8c))
* **deps:** update module github.com/prometheus/common to v0.70.1 ([#37](https://github.com/SpechtLabs/rpi_exporter/issues/37)) ([ef44207](https://github.com/SpechtLabs/rpi_exporter/commit/ef442077a07faff54b3e6253313a0f34c85bec75))
* **deps:** update module github.com/prometheus/common to v0.72.0 ([#46](https://github.com/SpechtLabs/rpi_exporter/issues/46)) ([8f62d70](https://github.com/SpechtLabs/rpi_exporter/commit/8f62d709aeaddcd9ee5ebeee623af62b323be157))
* **deps:** update module github.com/sirupsen/logrus to v1.10.0 ([#39](https://github.com/SpechtLabs/rpi_exporter/issues/39)) ([76a32e6](https://github.com/SpechtLabs/rpi_exporter/commit/76a32e644d0a497526cc3842a27fa0ef5f215f8e))
* **deps:** update module github.com/sirupsen/logrus to v1.10.1 ([#41](https://github.com/SpechtLabs/rpi_exporter/issues/41)) ([39ecd87](https://github.com/SpechtLabs/rpi_exporter/commit/39ecd8734cfb4a033cf9d9f3fc25ffc262d8da15))
* **deps:** update module github.com/sirupsen/logrus to v1.10.2 ([#43](https://github.com/SpechtLabs/rpi_exporter/issues/43)) ([3c9422f](https://github.com/SpechtLabs/rpi_exporter/commit/3c9422f3c3918aabb55f705eab3e01201b2f5c89))
* guard the filtered handler cache against concurrent scrapes ([9b23910](https://github.com/SpechtLabs/rpi_exporter/commit/9b23910deef446360327f7b20dd61d2b53765169))
* report fan collector failures to the scrape instead of exiting the exporter ([9b23910](https://github.com/SpechtLabs/rpi_exporter/commit/9b23910deef446360327f7b20dd61d2b53765169))
* stop the textfile collector from panicking on every .prom file ([9b23910](https://github.com/SpechtLabs/rpi_exporter/commit/9b23910deef446360327f7b20dd61d2b53765169))
