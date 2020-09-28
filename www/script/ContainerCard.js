export var ContainerCard = {
    props: ['container'],
    template: `
<div class="card-body">
    <div class="row align-items-center">
        <div class="col">
            <div class="row">
                <div class="col text-muted">Name</div>
                <div class="col text-muted">Image</div>
                <div class="col text-muted">State</div>
                <div class="col text-muted">Status</div>
            </div>
            <div class="row">
                <div class="col">{{ container.name }}</div>
                <div class="col">{{ container.image }}</div>
                <div class="col">
                    <div class="badge" :class="stateClass(container.state)">
                        {{ container.state }}
                    </div>
                </div>
                <div class="col">{{ container.status }}</div>
            </div>
        </div>
        <div class="col-auto">
            <div class="row text-right text-nowrap">
                <div class="dropdown">
                    <button class="btn btn-outline-primary btn-sm dropdown-toggle" type="button" id="actionDropDown" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false"></button>
                    <div class="dropdown-menu dropdown-menu-right" aria-labelledby="actionDropDown">
                        <a class="dropdown-item" :href="container.links.downloadLog.href" :class="container.links.downloadLog ? '' : 'disabled'" download>
                            <svg width="1em" height="1em" viewBox="0 0 16 16" class="bi bi-download mb-1 mr-2" fill="currentColor" xmlns="http://www.w3.org/2000/svg">
                                <path fill-rule="evenodd" d="M.5 9.9a.5.5 0 0 1 .5.5v2.5a1 1 0 0 0 1 1h12a1 1 0 0 0 1-1v-2.5a.5.5 0 0 1 1 0v2.5a2 2 0 0 1-2 2H2a2 2 0 0 1-2-2v-2.5a.5.5 0 0 1 .5-.5z" />
                                <path fill-rule="evenodd" d="M7.646 11.854a.5.5 0 0 0 .708 0l3-3a.5.5 0 0 0-.708-.708L8.5 10.293V1.5a.5.5 0 0 0-1 0v8.793L5.354 8.146a.5.5 0 1 0-.708.708l3 3z" />
                            </svg>
                            Download log
                        </a>
                        <a class="dropdown-item" @click="$emit('action', container.links.remove)" :class="container.links.remove ? '' : 'disabled'" href="#">
                            <svg width="1em" height="1em" viewBox="0 0 16 16" class="bi bi-trash mb-1 mr-2" fill="currentColor" xmlns="http://www.w3.org/2000/svg">
                                <path d="M5.5 5.5A.5.5 0 0 1 6 6v6a.5.5 0 0 1-1 0V6a.5.5 0 0 1 .5-.5zm2.5 0a.5.5 0 0 1 .5.5v6a.5.5 0 0 1-1 0V6a.5.5 0 0 1 .5-.5zm3 .5a.5.5 0 0 0-1 0v6a.5.5 0 0 0 1 0V6z"/>
                                <path fill-rule="evenodd" d="M14.5 3a1 1 0 0 1-1 1H13v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V4h-.5a1 1 0 0 1-1-1V2a1 1 0 0 1 1-1H6a1 1 0 0 1 1-1h2a1 1 0 0 1 1 1h3.5a1 1 0 0 1 1 1v1zM4.118 4L4 4.059V13a1 1 0 0 0 1 1h6a1 1 0 0 0 1-1V4.059L11.882 4H4.118zM2.5 3V2h11v1h-11z"/>
                            </svg>
                            Remove
                        </a>
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>
    `,
    methods: {
        stateClass: function (state) {
            switch (state) {
                case 'running':
                    return 'badge-success';
                case 'exited':
                    return 'badge-danger';
                default:
                    break;
            }
        }
    }
}