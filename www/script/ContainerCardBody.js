export var ContainerCardBody = {
    props: ['id', 'name', 'dataParent'],
    template: `
        <div :id="name" class="collapse" :aria-labelledby="id"
            :data-parent="dataParentHashed">
            <div class="card-body">
                Body
            </div>
        </div>
    `,
    computed: {
        dataParentHashed: function () {
            return '#' + this.dataParent;
        },
    }
}