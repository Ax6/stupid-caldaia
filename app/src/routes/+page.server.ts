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
	rules: Rule[];
};

export type Rule = {
	id: string;
	start: string;
	duration: string;
	targetTemp: number;
	repeatDays: [number];
	stoppedTime: Date;
	isActive: boolean;
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
				rules {
					id
					start
					duration
					targetTemp
					repeatDays
					stoppedTime
					isActive
				}
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
