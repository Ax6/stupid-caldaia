import { gql, madonna } from '$lib/porca-madonna-ql';

export type PageData = BoilerData & SensorData & SensorRangeData;

export type BoilerData = {
	boiler: Boiler;
};

export type SensorData = {
	sensor?: Measure;
};

export type SensorRangeData = {
	sensorRange: Measure[];
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
			sensor(name: "temperatura", position: "centrale") {
				timestamp
				value
			}
			sensorRange(name: "temperatura", position: "centrale") {
				timestamp
				value
			}
		}
	`);
}
