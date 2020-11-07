Vue.component("task", {
	props: ["data"],
	methods: {
		task_del: function () {
			this.$emit('task_del');
		},
		task_edit: function () {
			this.$emit('task_edit');
		},
		task_done: function () {
			this.$emit('task_done');
		},
		save: function () {
			this.$emit('save');
		},
		disable: function () {
			this.$emit('disable');
		},
	},
	template: `
	<div class="task" v-bind:class="{ taskFalse: data.checked}" >
		<div class="content">
			<div v-if="!data.editable" class="contentText">
				<p class="task_content">{{data.text}}</p>
				<div class="button check">
					<button v-if="!data.checked" @click="task_done()" class="task_done taskButton" style="color: rgba(26,205,30,0.89); font-size: 32px"> ☐ </button>
					<button v-if="data.checked" @click="task_done()" class="task_done taskButton" style="color: rgba(26,205,30,0.89); font-size: 32px"> ☑ </button>
					<button @click="task_edit()" class="task_edit taskButton" style="color: #eca81a; font-size: 30px"> ✎️ </button>
					<button @click="task_del()" class="task_del taskButton" style="color: #cd1537; font-size: 26px"> ✕ </button>
				</div>
			</div>
			<div v-if="data.editable">
       			<input v-on:keyup.enter="save()" v-model="data.inputedit" class="input" size="70px"/>
   				<button @click="save()">💾</button>
   				<button @click="disable()">⊗</button>
       		</div>	
		</div>
	</div>
	`
});

const url = "http://127.0.0.1:3000/";
const site = "http://127.0.0.1:3000/login.html";
let vue = new Vue({
	el: '#app',
	data: {
		new_task: {
			text: '',
			editable: false,
			checked: false
		},
		items: [{
			inputedit: '',
			checked: false,
		}]
	},
	methods: {
		getItems: function(){
			fetch(url + 'getItems', {credentials: 'include'})
				.then(res => res.json())
				.then((response) => {
					this.items = response.items.map((item) => {
						item.editable = false;
						return item;
					})
				});
		},
		getDelete: function(index){
			let request = JSON.stringify({id: index, });
			fetch(url + 'deleteItem', {
				method: 'DELETE',
				body: request,
				credentials: 'include',
				headers: {
					'Content-Type': 'application/json;'
				},
			}).then(res => res.json())
				.then((response) => {
					if(response['ok'] === true){
						this.getItems()
					} else {
						alert("Error 500. Internal server error. Please try again later")
					}
				});
		},
		getPost: function(){
			let request = JSON.stringify({text: this.new_task.text});
			fetch(url + 'addItem', {
				method: 'POST',
				body: request,
				credentials: 'include',
				headers: {
					'Content-Type': 'application/json;'
				},
			}).then(res => res.json())
				.then((response) => {
					if (response.id) {
						this.getItems();
						this.new_task.text = '';
					} else {
						alert("Error 500. Internal server error. Please try again later")
					}
				});
		},
		getPut: function(index, id){
			let request = JSON.stringify({text: this.items[index].text, id: id,  checked: this.items[index].checked});
			fetch(url + 'changeItem', {
				method: 'PUT',
				body: request,
				credentials: 'include',
				headers: {
					'Content-Type': 'application/json;'
				},
			}).then(res => res.json())
				.then((response) => {
				this.getItems()
			});
		},
		del(index){
			this.getDelete(index)
		},
		add_task() {
			if(this.new_task.text.trim() !== ''){
				this.getPost()
			}
		},
		task_done(index, id){
			this.items[index].checked = this.items[index].checked === false;
			this.checked = this.items[index].checked;
			this.getPut(index, id)
		},
		task_edit(index){
			this.items[index].inputedit = this.items[index].text;
			this.items[index].editable = true;
		},
		save(index, id){
			if(this.new_task.text !== '' || this.new_task.text !== ' ') {
				this.items[index].text = this.items[index].inputedit;
				this.getPut(index, id);
				this.items[index].editable = false;
			}},
		disable(index){
			this.items[index].editable = false;
			this.items[index].inputedit = '';
		},
		exit(){
			fetch(url + 'logout', {
				method: 'POST',
				credentials: 'include',
			}).then(res => res.json())
				.then((response) => {
					if(response.ok){
						localStorage.clear();
						window.location = site;
					}
				});
		}
	 },
	mounted() {
		this.getItems()
		},
	});
