<script lang="ts">
	import type { SensorRangeData, BoilerData, Measure } from './+page.server';
	import * as d3 from 'd3';

	export let data: SensorRangeData & BoilerData;

	let hoverData: any;
	let isTooltipHidden = true;
	let width = 0;
	let height = 350;

	const margin = { top: 20, right: 20, bottom: 20, left: 60 };
	const innerHeight = height - margin.top - margin.bottom;
	$: innerWidth = width - margin.left - margin.right;

	$: xValues = data.sensorRange.map((d) => new Date(d.timestamp));
	$: xDomain = [d3.min(xValues) || 0, d3.max(xValues) || 0];
	$: xScale = d3.scaleLinear().domain(xDomain).range([0, innerWidth]);

	$: yValues = data.sensorRange.map((d) => d.value);
	$: yDomain = [data.boiler.maxTemp, data.boiler.minTemp];
	$: yScale = d3.scaleLinear().domain(yDomain).range([0, innerHeight]);

	function hideTooltip() {
		isTooltipHidden = true;
	}

	function showTooltip(e: null | MouseEvent = null, d: null | Measure = null) {
		// Find closest point to e
		if (!d) {
			const mouseRelToSvg = d3.pointer(e);
			const mouseRelToPlot = [mouseRelToSvg[0] - margin.left, mouseRelToSvg[1] - margin.top];
			const x0 = new Date(xScale.invert(mouseRelToPlot[0]));
			const i = d3.bisect(xValues, x0, 1);
			const d0 = data.sensorRange[i - 1];
			const d1 = data.sensorRange[i];
			if (!d0 || !d1) return;
			d = x0 - new Date(d0.timestamp) > new Date(d1.timestamp) - x0 ? d1 : d0;
			if (Math.abs(yScale(d.value) - mouseRelToPlot[1]) > 20) return;
			if (Math.abs(xScale(new Date(d.timestamp)) - mouseRelToPlot[0]) > 20) return;
		}
		hoverData = {
			value: d.value,
			time: new Date(d.timestamp),
			x: xScale(new Date(d.timestamp)),
			y: yScale(d.value)
		};
		isTooltipHidden = false;
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
		<svg
			{width}
			{height}
			role="figure"
			on:mousemove={(e) => showTooltip(e)}
			on:mouseout={hideTooltip}
			on:blur={hideTooltip}
		>
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
					d={d3.line()(xValues.map((d, i) => [xScale(d), yScale(yValues[i])]))}
				>
				</path>
				<g>
					{#each data.sensorRange as d}
						<circle
							role="button"
							tabindex="0"
							cx={xScale(new Date(d.timestamp))}
							cy={yScale(d.value)}
							r="3"
							class="fill-slate-700"
							on:focus={() => showTooltip(undefined, d)}
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
				{#if !isTooltipHidden}
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
