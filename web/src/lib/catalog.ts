/**
 * Frontend mirror of the compiled-in inspection item catalog.
 * Contains required "shall describe" fields per system and short tab labels.
 */

export interface DescriptionField {
	key: string;
	label: string;
}

/** Required description fields per system, in SOP order. */
export const systemDescriptions: Record<string, DescriptionField[]> = {
	roof: [{ key: 'roof.covering_material', label: 'Type of roof-covering material' }],

	exterior: [{ key: 'exterior.wall_covering', label: 'Exterior wall covering material' }],

	foundation: [
		{ key: 'foundation.type', label: 'Type of foundation' },
		{ key: 'foundation.crawl_space_access', label: 'Crawl space access' }
	],

	heating: [
		{ key: 'heating.thermostat_location', label: 'Thermostat location' },
		{ key: 'heating.energy_source', label: 'Energy source' },
		{ key: 'heating.method', label: 'Heating method' }
	],

	cooling: [
		{ key: 'cooling.thermostat_location', label: 'Thermostat location' },
		{ key: 'cooling.method', label: 'Cooling method' }
	],

	plumbing: [
		{ key: 'plumbing.water_supply_type', label: 'Water supply type' },
		{ key: 'plumbing.main_water_shutoff_location', label: 'Main water shutoff location' },
		{ key: 'plumbing.main_fuel_shutoff_location', label: 'Main fuel shutoff location' },
		{ key: 'plumbing.fuel_storage_location', label: 'Fuel storage location' },
		{ key: 'plumbing.water_heater_capacity', label: 'Water heater capacity' }
	],

	electrical: [
		{ key: 'electrical.main_service_amperage', label: 'Main service amperage' },
		{ key: 'electrical.wiring_type', label: 'Wiring type' }
	],

	fireplace: [{ key: 'fireplace.type', label: 'Type of fireplace / fuel-burning appliance' }],

	attic: [
		{ key: 'attic.insulation_type', label: 'Insulation type' },
		{ key: 'attic.insulation_depth', label: 'Insulation depth' }
	],

	interior: [{ key: 'interior.garage_door_type', label: 'Garage door type' }]
};

/** Short labels for the system tab bar. */
export const systemShortLabel: Record<string, string> = {
	roof: 'Roof',
	exterior: 'Exterior',
	foundation: 'Foundation',
	heating: 'Heating',
	cooling: 'Cooling',
	plumbing: 'Plumbing',
	electrical: 'Electrical',
	fireplace: 'Fireplace',
	attic: 'Attic',
	interior: 'Interior'
};

export const allSystemTypes = [
	'roof',
	'exterior',
	'foundation',
	'heating',
	'cooling',
	'plumbing',
	'electrical',
	'fireplace',
	'attic',
	'interior'
] as const;
