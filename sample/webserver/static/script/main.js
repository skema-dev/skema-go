'use strict';

class Lesson extends React.Component {
    constructor(props) {
        console.log(props)
        super(props);
        this.state = {
            lessonId: props.lessonId,
            name: props.name,
            description: props.description
        }
    }

    loadLession() {
        let lessons = new Map()
        lessons.set("lesson1", <Lesson1 />)
        lessons.set("lesson2", <Lesson2 />)
        lessons.set("lesson3", <Lesson3 />)
        root.render(lessons.get(this.state.name));
    }

    render() {
        return (
            <div>
                <button onClick={() => this.loadLession() }>
                    { this.state.name }
                </button>
                <br/>
                { this.state.description }
            </div>
        );
    }
}

class ApiList extends React.Component {
    constructor(props) {
        console.log(props)
        super(props);
        this.state = {
            result: ""
        }
        this.healthCheck = this.healthCheck.bind(this);
        this.helloWorld = this.helloWorld.bind(this);
    }

    healthCheck() {
        axios
            .get('/api/healthcheck')
            .then(response => {
                this.setState({result: "healthcheck: " + response.data.result});
            })
            .catch(function (error) { // 请求失败处理
                console.log(error);
            });
    }

    helloWorld() {
        axios
            .post('/api/helloworld')
            .then(response => {
                this.setState({result: "helloworld:" + response.data.msg});
            })
            .catch(function (error) { // 请求失败处理
                console.log(error);
            });
    }

    render() {
        return (
            <div style={{ textAlign: "center" }}>
                <b>API List</b>
                <hr/>
                <button onClick={this.healthCheck}>
                    Call HealthCheck
                </button>
                <br/>
                <br/>
                <button onClick={this.helloWorld}>
                    Call HelloWorld
                </button>
                <br/>
                <br/>
                { this.state.result }
            </div>
        );
    }
}


const domContainer1 = document.querySelector('#main_container');
const root = ReactDOM.createRoot(domContainer1);
root.render(<ApiList />);
