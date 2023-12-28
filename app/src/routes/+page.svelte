<script lang="ts">
    import { graphql } from '$houdini';

    const temperatura = graphql(`
        subscription ItemUpdate($position: String!) {
            onTemperaturaChange(position: "centrale")
        }
    `)
</script>


<section>
    <div>
        <h1>Temperatura attuale</h1>
    </div>
</section>