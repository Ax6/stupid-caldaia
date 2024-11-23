import { fetchWeatherApi } from 'openmeteo';
import { LATITUDE, LONGITUDE, TIMEZONE } from '$env/static/private'

export type WeatherSample = { time: Date; temperature2m: number };

const params = {
    "timezone": TIMEZONE,
    "latitude": LATITUDE,
    "longitude": LONGITUDE,
    "past_days": 1,
    "hourly": "temperature_2m",
};

const url = "https://api.open-meteo.com/v1/forecast";

export async function getOutsideTemperature() {
    const responses = await fetchWeatherApi(url, params);

    // Helper function to form time ranges
    const range = (start: number, stop: number, step: number) =>
        Array.from({ length: (stop - start) / step }, (_, i) => start + i * step);

    // Process first location. Add a for-loop for multiple locations or weather models
    const response = responses[0];

    // Attributes for timezone and location
    const utcOffsetSeconds = response.utcOffsetSeconds();
    const hourly = response.hourly()!;
    const weatherData = {
        hourly: {
            time: range(Number(hourly.time()), Number(hourly.timeEnd()), hourly.interval()).map(
                (t) => new Date((t + utcOffsetSeconds) * 1000)
            ),
            temperature2m: hourly.variables(0)!.valuesArray()!,
        },
    };

    // `weatherData` now contains a simple structure with arrays for datetime and weather data
    let samples: WeatherSample[] = [];
    for (let i = 0; i < weatherData.hourly.time.length; i++) {
        samples.push({
            time: weatherData.hourly.time[i],
            temperature2m: weatherData.hourly.temperature2m[i]
        });
    }
    return samples;
}