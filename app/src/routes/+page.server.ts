import type { PageData } from '$lib/types';
import { gql, madonna } from '$lib/porca-madonna-ql';


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
