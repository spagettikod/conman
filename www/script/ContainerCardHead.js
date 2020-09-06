export var ContainerCardHead = {
    props: ['id', 'name'],
    template: `
        <div class="card-header" :id="id">
            <h2 class="mb-0">
                <button class="btn btn-link btn-block text-left" type="button" data-toggle="collapse"
                    :data-target="dataTarget" aria-expanded="false"
                    :aria-controls="name">
                    {{ name }}
                </button>
            </h2>
        </div>
    `,
    computed: {
        dataTarget: function () {
            return '#' + this.name;
        },
    }
}