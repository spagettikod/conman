import { ContainerCardHead } from './ContainerCardHead.js'
import { ContainerCardBody } from './ContainerCardBody.js'

export var ContainerCard = {
    props: ['container', 'dataParent'],
    template: `
        <div class="card">
            <card-head :id="container.id" :name="container.name"></card-head>
            <card-body :id="container.id" :name="container.name" :data-parent="dataParent"></card-body>
        </div>
    `,
    components: {
        'card-head': ContainerCardHead,
        'card-body': ContainerCardBody
    },
}