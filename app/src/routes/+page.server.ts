import type { PageData, LocalData, ExternalData, Measure } from '$lib/types';
import { gql, madonna } from '$lib/porca-madonna-ql';
import { getOutsideTemperature } from '$lib/outside';

async function getLocalData(): Promise<LocalData> {
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
					delay
					targetTemp
					repeatDays
					stoppedTime
					isActive
				}
			}
			currentTemperature: sensor(name: "temperatura", position: "centrale") {
				timestamp
				value
			}
			temperatureSeries: sensorRange(name: "temperatura", position: "centrale") {
				timestamp
				value
			}
			humiditySeries: sensorRange(name: "umidita", position: "centrale") {
				timestamp
				value
			}
		}
	`);
}

async function getExternalData(): Promise<ExternalData> {
	const weatherData = await getOutsideTemperature();
	const outsideTemperature = weatherData.map(
		(sample) =>
			({
				timestamp: sample.time.toISOString(),
				value: sample.temperature2m
			}) as Measure
	);
	return {
		outsideTemperatureSeries: outsideTemperature
	} as ExternalData;
}

export async function load(): Promise<PageData> {
	const reqLocal = getLocalData();
	const reqExternal = getExternalData();
	const localData = await reqLocal;
	const externalData = await reqExternal;
	return {
		...localData,
		...externalData
	};
}
