<script lang="ts">
	import type { Measure } from '$lib/types';
	import { pan, type GestureCustomEvent } from 'svelte-gestures';
	import * as d3 from 'd3';

	interface Props {
		data: Measure[];
		title: string;
		yLabel: string;
		height?: number;
	}

	let { data, title, yLabel, height = 350 }: Props = $props();

	let hoverData: any = $state();
	let isTooltipHidden = $state(true);
	let width = $state(0);

	const margin = { top: 0, right: 20, bottom: 20, left: 60 };
	const innerHeight = height - margin.top - margin.bottom;
	let innerWidth = $derived(width - margin.left - margin.right);

	let xValues = $derived(data.map((d) => new Date(d.timestamp)));
	let xDomain = $derived([d3.min(xValues) || new Date(), d3.max(xValues) || new Date()]);
	let xScale = $derived(d3.scaleLinear().domain(xDomain).range([0, innerWidth]));

	let xTicks = $derived(
		d3.range(0, 24, Math.round(1800 / width)).reduce((acc: Date[], curr: number) => {
			const delta = xDomain[1].getTime() - 1000 * 60 * 60 * curr;
			const closestHour = new Date(delta);
			closestHour.setMinutes(0);
			return closestHour.getTime() > xDomain[0].getTime() ? [...acc, closestHour] : acc;
		}, [])
	);

	let yValues = $derived(data.map((d) => d.value));
	let yMinMax = $derived([d3.max(yValues) || 1, d3.min(yValues) || -1]);
	let yMargin = $derived((yMinMax[1] - yMinMax[0]) / 20);
	let yDomain = $derived([yMinMax[0] - yMargin, yMinMax[1] + yMargin]);
	let yScale = $derived(d3.scaleLinear().domain(yDomain).range([0, innerHeight]));

	function getHoveredSample(x: number, y: number): Measure {
		const mouseRelToPlot = [x - margin.left, y - margin.top];
		const x0 = new Date(xScale.invert(mouseRelToPlot[0]));
		const i = d3.bisect(xValues, x0, 1);
		const d0 = data[i - 1];
		const d1 = data[i];
		if (!d0 || !d1)
			return {
				timestamp: '0',
				value: 0
			} as Measure;
		const distanceToPrevious = x0.getTime() - new Date(d0.timestamp).getTime();
		const distanceToFollowing = new Date(d1.timestamp).getTime() - x0.getTime();
		return distanceToPrevious > distanceToFollowing ? d1 : d0;
	}

	function hideTooltip() {
		isTooltipHidden = true;
	}

	function showTooltip(d: Measure) {
		hoverData = {
			value: d.value,
			time: new Date(d.timestamp),
			x: xScale(new Date(d.timestamp)),
			y: yScale(d.value)
		};
		isTooltipHidden = false;
	}
</script>

<div class="w-full bg-gray-200 border border-gray-300 rounded-xl" bind:clientWidth={width}>
	<div class="m-2">
		<h1 class="text-2xl font-thin">{title}</h1>
	</div>
	{#if width === 0}
		<p class="text-2xl m-2 font-semibold">Caricamento...</p>
	{:else}
		<div
			role="figure"
			onmousemove={(e) => {
				const [x, y] = d3.pointer(e);
				showTooltip(getHoveredSample(x, y));
			}}
			onmouseout={hideTooltip}
			onblur={hideTooltip}
			use:pan
			onpanmove={(e) => {
				showTooltip(getHoveredSample(e.detail.x, e.detail.y));
			}}
			onpanup={hideTooltip}
		>
			<svg {width} {height} class="overflow-visible">
				<g transform={`translate(${margin.left},${margin.top})`}>
					{#each xTicks as tickValue}
						<g transform={`translate(${xScale(tickValue)},0)`}>
							<line y2={innerHeight} class="stroke-gray-400" />
							<text text-anchor="middle" dy=".71em" y={innerHeight + 3}>
								{tickValue.getHours()}:{tickValue.getMinutes() < 10 ? '00' : '30'}
							</text>
						</g>
					{/each}
					{#each yScale.ticks(5) as tickValue}
						<g transform={`translate(0,${yScale(tickValue)})`}>
							<line x2={innerWidth} class="stroke-gray-400" />
							<text text-anchor="end" x="-3" dy=".32em">
								{tickValue}
							</text>
						</g>
					{/each}
					<path
						class="stroke-slate-800 fill-none"
						stroke-width="1"
						d={d3.line()(xValues.map((d, i) => [xScale(d), yScale(yValues[i])]))}
					>
					</path>
					<g>
						{#each data as d}
							<circle
								role="button"
								tabindex="0"
								r="1"
								cx={xScale(new Date(d.timestamp))}
								cy={yScale(d.value)}
								class="fill-slate-700"
								onfocus={() => showTooltip(d)}
								onblur={hideTooltip}
							/>
						{/each}
					</g>
					<text
						text-anchor="middle"
						transform={`translate(${-margin.left / 2 - margin.left / 10},${
							innerHeight / 2
						}) rotate(-90)`}
					>
						{yLabel}
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
								y={hoverData.y - yScale.invert(yDomain[1]) < 30
									? hoverData.y + 20
									: hoverData.y - 50}
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
		</div>
	{/if}
</div>
