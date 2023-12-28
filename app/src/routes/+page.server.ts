import { gql, madonna } from '$lib/porca-madonna-ql';

export type PageData = BoilerData & TemperatureData & TemperatureRangeData;

export type BoilerData = {
	boiler: Boiler;
};

export type TemperatureData = {
	temperature?: Measure;
};

export type TemperatureRangeData = {
	temperatureRange: Measure[];
};

export type Boiler = {
	state: State;
	minTemp: number;
	maxTemp: number;
	targetTemp: number;
};

export type State = 'OFF' | 'ON';

export type Measure = {
	timestamp: string;
	value: number;
};

export async function load(): Promise<PageData> {
	return await madonna(gql`
		query {
			boiler {
				state
				minTemp
				maxTemp
				targetTemp
			}
			temperature(position: "centrale") {
				timestamp
				value
			}
			temperatureRange(position: "centrale") {
				timestamp
				value
			}
		}
	`);
}
