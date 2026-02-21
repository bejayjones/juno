package domain

// ItemDefinition is a compiled-in catalog entry for an inspectable item.
type ItemDefinition struct {
	Key   ItemKey
	Label string
}

// DescriptionRequirement defines a required "shall describe" field for a system.
type DescriptionRequirement struct {
	Key   DescriptionKey
	Label string // Human-readable prompt shown to the inspector.
}

// SystemDefinition is the compiled-in definition of one InterNACHI inspection system.
type SystemDefinition struct {
	Type                 SystemType
	Label                string
	RequiredDescriptions []DescriptionRequirement
	Items                []ItemDefinition
}

// DescriptionKey constants — one per required "shall describe" SOP obligation.
const (
	DescRoofCoveringMaterial     DescriptionKey = "roof.covering_material"
	DescExteriorWallCovering     DescriptionKey = "exterior.wall_covering"
	DescFoundationType           DescriptionKey = "foundation.type"
	DescCrawlSpaceAccess         DescriptionKey = "foundation.crawl_space_access"
	DescHeatingThermostatLoc     DescriptionKey = "heating.thermostat_location"
	DescHeatingEnergySource      DescriptionKey = "heating.energy_source"
	DescHeatingMethod            DescriptionKey = "heating.method"
	DescCoolingThermostatLoc     DescriptionKey = "cooling.thermostat_location"
	DescCoolingMethod            DescriptionKey = "cooling.method"
	DescWaterSupplyType          DescriptionKey = "plumbing.water_supply_type"
	DescMainWaterShutoffLocation DescriptionKey = "plumbing.main_water_shutoff_location"
	DescMainFuelShutoffLocation  DescriptionKey = "plumbing.main_fuel_shutoff_location"
	DescFuelStorageLocation      DescriptionKey = "plumbing.fuel_storage_location"
	DescWaterHeaterCapacity      DescriptionKey = "plumbing.water_heater_capacity"
	DescMainServiceAmperage      DescriptionKey = "electrical.main_service_amperage"
	DescWiringType               DescriptionKey = "electrical.wiring_type"
	DescFireplaceType            DescriptionKey = "fireplace.type"
	DescInsulationType           DescriptionKey = "attic.insulation_type"
	DescInsulationDepth          DescriptionKey = "attic.insulation_depth"
	DescGarageDoorType           DescriptionKey = "interior.garage_door_type"
)

// ItemKey constants — one per inspectable item across all ten systems.
const (
	// Roof (SOP 3.1)
	ItemGuttersDownspouts    ItemKey = "roof.gutters_downspouts"
	ItemVents                ItemKey = "roof.vents"
	ItemRoofFlashing         ItemKey = "roof.flashing"
	ItemSkylights            ItemKey = "roof.skylights"
	ItemChimneys             ItemKey = "roof.chimneys"
	ItemRoofPenetrations     ItemKey = "roof.penetrations"
	ItemGeneralRoofStructure ItemKey = "roof.general_structure"

	// Exterior (SOP 3.2)
	ItemEavesSoffitsFascia     ItemKey = "exterior.eaves_soffits_fascia"
	ItemExteriorWindows        ItemKey = "exterior.windows"
	ItemExteriorDoors          ItemKey = "exterior.doors"
	ItemExteriorFlashingTrim   ItemKey = "exterior.flashing_trim"
	ItemWalkwaysDriveways      ItemKey = "exterior.walkways_driveways"
	ItemExteriorStairs         ItemKey = "exterior.stairs_steps_ramps"
	ItemPorchesPatiosDecks     ItemKey = "exterior.porches_patios_decks"
	ItemExteriorRailingsGuards ItemKey = "exterior.railings_guards"
	ItemVegetationDrainage     ItemKey = "exterior.vegetation_drainage"

	// Foundation / Basement / Crawl Space (SOP 3.3)
	ItemFoundation           ItemKey = "foundation.foundation"
	ItemBasement             ItemKey = "foundation.basement"
	ItemCrawlSpace           ItemKey = "foundation.crawl_space"
	ItemStructuralComponents ItemKey = "foundation.structural_components"

	// Heating (SOP 3.4)
	ItemHeatingSystem ItemKey = "heating.system"

	// Cooling (SOP 3.5)
	ItemCoolingSystem ItemKey = "cooling.system"

	// Plumbing (SOP 3.6)
	ItemMainWaterShutoff    ItemKey = "plumbing.main_water_shutoff"
	ItemMainFuelShutoff     ItemKey = "plumbing.main_fuel_shutoff"
	ItemWaterHeater         ItemKey = "plumbing.water_heater"
	ItemInteriorWaterSupply ItemKey = "plumbing.interior_water_supply"
	ItemToilets             ItemKey = "plumbing.toilets"
	ItemSinksTubsShowers    ItemKey = "plumbing.sinks_tubs_showers"
	ItemDrainWasteVent      ItemKey = "plumbing.drain_waste_vent"
	ItemSumpPumps           ItemKey = "plumbing.sump_pumps"

	// Electrical (SOP 3.7)
	ItemServiceDrop                 ItemKey = "electrical.service_drop"
	ItemServiceHead                 ItemKey = "electrical.service_head"
	ItemElectricMeter               ItemKey = "electrical.meter"
	ItemMainServiceDisconnect       ItemKey = "electrical.main_service_disconnect"
	ItemPanelboards                 ItemKey = "electrical.panelboards"
	ItemServiceGroundingBonding     ItemKey = "electrical.grounding_bonding"
	ItemSwitchesLightingReceptacles ItemKey = "electrical.switches_lighting_receptacles"
	ItemAFCIProtection              ItemKey = "electrical.afci_protection"
	ItemGFCIProtection              ItemKey = "electrical.gfci_protection"
	ItemSmokeDetectors              ItemKey = "electrical.smoke_detectors"
	ItemCarbonMonoxideDetectors     ItemKey = "electrical.carbon_monoxide_detectors"

	// Fireplace (SOP 3.8)
	ItemFireplaceChimney ItemKey = "fireplace.chimney"
	ItemLintels          ItemKey = "fireplace.lintels"
	ItemDamperDoors      ItemKey = "fireplace.damper_doors"
	ItemCleanoutDoors    ItemKey = "fireplace.cleanout_doors"

	// Attic / Insulation / Ventilation (SOP 3.9)
	ItemAtticInsulation   ItemKey = "attic.insulation"
	ItemAtticVentilation  ItemKey = "attic.ventilation"
	ItemKitchenExhaust    ItemKey = "attic.kitchen_exhaust"
	ItemBathroomExhaust   ItemKey = "attic.bathroom_exhaust"
	ItemLaundryExhaust    ItemKey = "attic.laundry_exhaust"

	// Doors / Windows / Interior (SOP 3.10)
	ItemInteriorDoors      ItemKey = "interior.doors"
	ItemInteriorWindows    ItemKey = "interior.windows"
	ItemFloors             ItemKey = "interior.floors"
	ItemWalls              ItemKey = "interior.walls"
	ItemCeilings           ItemKey = "interior.ceilings"
	ItemInteriorStairs     ItemKey = "interior.stairs"
	ItemInteriorRailings   ItemKey = "interior.railings"
	ItemGarageVehicleDoors ItemKey = "interior.garage_vehicle_doors"
	ItemGarageDoorOpeners  ItemKey = "interior.garage_door_openers"
)

// Catalog is the complete, compiled-in InterNACHI inspection item catalog.
// It defines all ten inspection systems in SOP order (3.1–3.10), their
// inspectable items, and their required "shall describe" description fields.
var Catalog = []*SystemDefinition{
	{
		Type:  SystemRoof,
		Label: "Roof",
		RequiredDescriptions: []DescriptionRequirement{
			{Key: DescRoofCoveringMaterial, Label: "Type of roof-covering material"},
		},
		Items: []ItemDefinition{
			{Key: ItemGuttersDownspouts, Label: "Gutters and downspouts"},
			{Key: ItemVents, Label: "Vents"},
			{Key: ItemRoofFlashing, Label: "Flashing"},
			{Key: ItemSkylights, Label: "Skylights"},
			{Key: ItemChimneys, Label: "Chimneys"},
			{Key: ItemRoofPenetrations, Label: "Roof penetrations"},
			{Key: ItemGeneralRoofStructure, Label: "General roof structure"},
		},
	},
	{
		Type:  SystemExterior,
		Label: "Exterior",
		RequiredDescriptions: []DescriptionRequirement{
			{Key: DescExteriorWallCovering, Label: "Type of exterior wall-covering material"},
		},
		Items: []ItemDefinition{
			{Key: ItemEavesSoffitsFascia, Label: "Eaves, soffits, and fascia"},
			{Key: ItemExteriorWindows, Label: "Windows (representative number)"},
			{Key: ItemExteriorDoors, Label: "Exterior doors"},
			{Key: ItemExteriorFlashingTrim, Label: "Flashing and trim"},
			{Key: ItemWalkwaysDriveways, Label: "Walkways and driveways"},
			{Key: ItemExteriorStairs, Label: "Stairs, steps, stoops, and ramps"},
			{Key: ItemPorchesPatiosDecks, Label: "Porches, patios, decks, balconies, and carports"},
			{Key: ItemExteriorRailingsGuards, Label: "Railings, guards, and handrails"},
			{Key: ItemVegetationDrainage, Label: "Vegetation, surface drainage, and grading"},
		},
	},
	{
		Type:  SystemFoundation,
		Label: "Foundation / Basement / Crawl Space",
		RequiredDescriptions: []DescriptionRequirement{
			{Key: DescFoundationType, Label: "Type of foundation"},
			{Key: DescCrawlSpaceAccess, Label: "Location of under-floor space access"},
		},
		Items: []ItemDefinition{
			{Key: ItemFoundation, Label: "Foundation"},
			{Key: ItemBasement, Label: "Basement"},
			{Key: ItemCrawlSpace, Label: "Crawl space"},
			{Key: ItemStructuralComponents, Label: "Structural components"},
		},
	},
	{
		Type:  SystemHeating,
		Label: "Heating",
		RequiredDescriptions: []DescriptionRequirement{
			{Key: DescHeatingThermostatLoc, Label: "Location of thermostat"},
			{Key: DescHeatingEnergySource, Label: "Energy source"},
			{Key: DescHeatingMethod, Label: "Heating method"},
		},
		Items: []ItemDefinition{
			{Key: ItemHeatingSystem, Label: "Heating system"},
		},
	},
	{
		Type:  SystemCooling,
		Label: "Cooling",
		RequiredDescriptions: []DescriptionRequirement{
			{Key: DescCoolingThermostatLoc, Label: "Location of thermostat"},
			{Key: DescCoolingMethod, Label: "Cooling method"},
		},
		Items: []ItemDefinition{
			{Key: ItemCoolingSystem, Label: "Cooling system"},
		},
	},
	{
		Type:  SystemPlumbing,
		Label: "Plumbing",
		RequiredDescriptions: []DescriptionRequirement{
			{Key: DescWaterSupplyType, Label: "Water supply type (public or private)"},
			{Key: DescMainWaterShutoffLocation, Label: "Location of main water supply shut-off valve"},
			{Key: DescMainFuelShutoffLocation, Label: "Location of main fuel supply shut-off valve"},
			{Key: DescFuelStorageLocation, Label: "Location of fuel storage systems"},
			{Key: DescWaterHeaterCapacity, Label: "Capacity of water heating equipment (if labeled)"},
		},
		Items: []ItemDefinition{
			{Key: ItemMainWaterShutoff, Label: "Main water supply shut-off valve"},
			{Key: ItemMainFuelShutoff, Label: "Main fuel supply shut-off valve"},
			{Key: ItemWaterHeater, Label: "Water heating equipment"},
			{Key: ItemInteriorWaterSupply, Label: "Interior water supply fixtures and faucets"},
			{Key: ItemToilets, Label: "Toilets"},
			{Key: ItemSinksTubsShowers, Label: "Sinks, tubs, and showers (functional drainage)"},
			{Key: ItemDrainWasteVent, Label: "Drain, waste, and vent system"},
			{Key: ItemSumpPumps, Label: "Drainage sump pumps"},
		},
	},
	{
		Type:  SystemElectrical,
		Label: "Electrical",
		RequiredDescriptions: []DescriptionRequirement{
			{Key: DescMainServiceAmperage, Label: "Main service amperage rating (if labeled)"},
			{Key: DescWiringType, Label: "Type of wiring observed"},
		},
		Items: []ItemDefinition{
			{Key: ItemServiceDrop, Label: "Service drop"},
			{Key: ItemServiceHead, Label: "Service head, gooseneck, and drip loops"},
			{Key: ItemElectricMeter, Label: "Electric meter and base"},
			{Key: ItemMainServiceDisconnect, Label: "Main service disconnect"},
			{Key: ItemPanelboards, Label: "Panelboards and over-current protection devices"},
			{Key: ItemServiceGroundingBonding, Label: "Service grounding and bonding"},
			{Key: ItemSwitchesLightingReceptacles, Label: "Switches, lighting fixtures, and receptacles (representative number)"},
			{Key: ItemAFCIProtection, Label: "AFCI-protected receptacles"},
			{Key: ItemGFCIProtection, Label: "GFCI receptacles and circuit breakers"},
			{Key: ItemSmokeDetectors, Label: "Smoke detectors"},
			{Key: ItemCarbonMonoxideDetectors, Label: "Carbon monoxide detectors"},
		},
	},
	{
		Type:  SystemFireplace,
		Label: "Fireplace",
		RequiredDescriptions: []DescriptionRequirement{
			{Key: DescFireplaceType, Label: "Type of fireplace"},
		},
		Items: []ItemDefinition{
			{Key: ItemFireplaceChimney, Label: "Fireplace and chimney (accessible and visible portions)"},
			{Key: ItemLintels, Label: "Lintels above fireplace openings"},
			{Key: ItemDamperDoors, Label: "Damper doors"},
			{Key: ItemCleanoutDoors, Label: "Cleanout doors and frames"},
		},
	},
	{
		Type:  SystemAttic,
		Label: "Attic / Insulation / Ventilation",
		RequiredDescriptions: []DescriptionRequirement{
			{Key: DescInsulationType, Label: "Type of insulation observed"},
			{Key: DescInsulationDepth, Label: "Approximate average depth of insulation at unfinished attic floor (inches)"},
		},
		Items: []ItemDefinition{
			{Key: ItemAtticInsulation, Label: "Insulation in unfinished spaces"},
			{Key: ItemAtticVentilation, Label: "Ventilation of unfinished spaces"},
			{Key: ItemKitchenExhaust, Label: "Kitchen mechanical exhaust"},
			{Key: ItemBathroomExhaust, Label: "Bathroom mechanical exhaust"},
			{Key: ItemLaundryExhaust, Label: "Laundry area mechanical exhaust"},
		},
	},
	{
		Type:  SystemInterior,
		Label: "Doors / Windows / Interior",
		RequiredDescriptions: []DescriptionRequirement{
			{Key: DescGarageDoorType, Label: "Garage vehicle door type (manually-operated or opener-equipped)"},
		},
		Items: []ItemDefinition{
			{Key: ItemInteriorDoors, Label: "Interior doors (representative number)"},
			{Key: ItemInteriorWindows, Label: "Interior windows (representative number)"},
			{Key: ItemFloors, Label: "Floors"},
			{Key: ItemWalls, Label: "Walls"},
			{Key: ItemCeilings, Label: "Ceilings"},
			{Key: ItemInteriorStairs, Label: "Stairs, steps, landings, and stairways"},
			{Key: ItemInteriorRailings, Label: "Railings, guards, and handrails"},
			{Key: ItemGarageVehicleDoors, Label: "Garage vehicle doors"},
			{Key: ItemGarageDoorOpeners, Label: "Garage vehicle door openers"},
		},
	},
}

// CatalogBySystem returns the SystemDefinition for the given SystemType,
// or nil if the type is not in the catalog.
func CatalogBySystem(t SystemType) *SystemDefinition {
	for _, def := range Catalog {
		if def.Type == t {
			return def
		}
	}
	return nil
}
