package entity

import (
	"encoding/json"
	"slices"
)

type ReagentType string

const (
	ReagentTypeVER     ReagentType = "VER"
	ReagentTypeCAL     ReagentType = "CAL"
	ReagentTypeALCOHOL ReagentType = "ALCOHOL"
	ReagentTypeBLEACH  ReagentType = "BLEACH"
	ReagentTypeWATER   ReagentType = "WATER"
	ReagentTypeALKALI  ReagentType = "ALKALI"
	ReagentTypeDF      ReagentType = "DF"
)

type Address string

const (
	AddressRA Address = "RA"
	AddressRB Address = "RB"
	AddressRC Address = "RC"
	AddressRD Address = "RD"
	AddressSA Address = "SA"
	AddressSB Address = "SB"
	AddressSC Address = "SC"
	AddressSD Address = "SD"
	AddressSE Address = "SE"
	AddressSF Address = "SF"
	AddressSG Address = "SG"
	AddressSH Address = "SH"
)

type Container struct {
	Address     Address
	ReagentType ReagentType
	Volume      int64
}
type Command struct {
	Entity
	Name             string
	SystemName       string
	Reagent          ReagentType
	AverageTime      int64
	VolumeWaste      int64
	VolumeDriveFluid int64
	VolumeContainer  int64
	DefaultAddress   Address
}

// Временная структура для парсинга
type CommandJSON struct {
	Name []struct {
		Locale string `json:"locale"`
		Value  string `json:"value"`
	} `json:"name"`
	SystemName       string  `json:"systemName"`
	Reagent          *string `json:"reagent"`
	AverageTime      int64   `json:"averageTime"`
	VolumeWaste      int64   `json:"volumeWaste"`
	VolumeDriveFluid int64   `json:"volumeDriveFluid"`
	VolumeContainer  int64   `json:"volumeContainer"`
	DefaultAddress   *string `json:"defaultAddress"`
}

func (c *Command) UnmarshalJSON(data []byte) error {
	var raw CommandJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	for _, n := range raw.Name {
		if n.Locale == "en" {
			c.Name = n.Value
			break
		}
	}
	if c.Name == "" {
		return ErrCommandNameNotFound
	}

	c.SystemName = raw.SystemName
	c.AverageTime = raw.AverageTime
	c.VolumeWaste = raw.VolumeWaste
	c.VolumeDriveFluid = raw.VolumeDriveFluid
	c.VolumeContainer = raw.VolumeContainer

	if raw.Reagent != nil {
		c.Reagent = ReagentType(*raw.Reagent)
	}
	if raw.DefaultAddress != nil {
		c.DefaultAddress = Address(*raw.DefaultAddress)
	}

	return nil
}

func (t Container) IsValidVolume() bool {
	one := []Address{AddressSA, AddressSB, AddressSC, AddressSD, AddressSE, AddressSF, AddressSG, AddressSH}
	two := []Address{AddressRA, AddressRB, AddressRC, AddressRD}

	if slices.Contains(one, t.Address) && t.Volume > 200 {
		return false
	}
	if slices.Contains(two, t.Address) && t.Volume > 5000 {
		return false
	}
	return true
}
