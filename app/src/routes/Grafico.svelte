<script lang="ts">
	import type { TemperatureRangeData } from './+page.server';
	import * as d3 from 'd3';

	export let data: TemperatureRangeData;

	let hoverData: any = { show: false };
	let width = 0;
	let height = 400;

	const margin = { top: 20, right: 20, bottom: 20, left: 60 };
	const innerHeight = height - margin.top - margin.bottom;
	$: innerWidth = width - margin.left - margin.right;

	$: xDomain = data.temperatureRange.map((d) => new Date(d.timestamp));
	$: yScale = d3.scaleLinear().domain([100, 0]).range([0, innerHeight]);
	$: xScale = d3
		.scaleLinear()
		.domain([d3.min(xDomain) || 0, d3.max(xDomain) || 0])
		.range([0, innerWidth]);

	function hideTooltip() {
		hoverData = { show: false };
	}

	function showTooltip(d: any) {
		hoverData = {
			value: d.value,
			time: new Date(d.timestamp),
			x: xScale(new Date(d.timestamp)),
			y: yScale(d.value),
			show: true
		};
	}
</script>

<div class="w-full" bind:clientWidth={width}>
	<div class="m-2">
		<h1 class="text-4xl font-bold">Grafico</h1>
		<p class="text-xl font-semibold">Temperatura</p>
	</div>
	{#if width === 0}
		<p class="text-2xl m-2 font-semibold">Caricamento...</p>
	{:else}
		<svg {width} {height}>
			<g transform={`translate(${margin.left},${margin.top})`}>
				{#each xScale.ticks(width / 100) as tickValue}
					<g transform={`translate(${xScale(tickValue)},0)`}>
						<line y2={innerHeight} stroke="black" />
						<text text-anchor="middle" dy=".71em" y={innerHeight + 3}>
							{new Date(tickValue).getHours()}:00
						</text>
					</g>
				{/each}
				{#each yScale.ticks() as tickValue}
					<g transform={`translate(0,${yScale(tickValue)})`}>
						<line x2={innerWidth} stroke="black" />
						<text text-anchor="end" x="-3" dy=".32em">
							{tickValue}
						</text>
					</g>
				{/each}
				<path
					class="stroke-slate-800 fill-none stroke-width-2"
					d={d3.line()(
						data.temperatureRange.map((d) => [xScale(new Date(d.timestamp)), yScale(d.value)])
					)}
				>
				</path>
				<g>
					{#each data.temperatureRange as d}
						<circle
							role="button"
							tabindex="0"
							cx={xScale(new Date(d.timestamp))}
							cy={yScale(d.value)}
							r="3"
							class="fill-slate-700"
							on:mouseenter={() => showTooltip(d)}
							on:focus={() => showTooltip(d)}
							on:mouseout={hideTooltip}
							on:blur={hideTooltip}
						/>
					{/each}
				</g>
				<text
					text-anchor="middle"
					transform={`translate(${-margin.left / 2 - margin.left / 10},${
						innerHeight / 2
					}) rotate(-90)`}
				>
					Temperatura (°C)
				</text>
				{#if hoverData.show}
					<g>
						<line
							x1={hoverData.x}
							y1={0}
							x2={hoverData.x}
							y2={innerHeight}
							stroke="black"
							stroke-dasharray="5,5"
						/>
						<line
							x1={0}
							y1={hoverData.y}
							x2={innerWidth}
							y2={hoverData.y}
							stroke="black"
							stroke-dasharray="5,5"
						/>
						<foreignObject
							x={hoverData.x + margin.left > width / 2
								? Math.min(hoverData.x - 75, width - 150 - margin.left - margin.right)
								: Math.max(hoverData.x - 75, 0)}
							y={hoverData.y - 50}
							width="150"
							height="40"
						>
							<div class="bg-slate-700 rounded p-1">
								<p class="text-white text-center">
									{hoverData.value.toFixed(2)}°C alle {hoverData.time
										.toTimeString()
										.split(' ')[0]
										.slice(0, -3)}
								</p>
							</div>
						</foreignObject>
					</g>
				{/if}
			</g>
		</svg>
	{/if}
</div>
