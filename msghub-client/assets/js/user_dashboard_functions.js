// To change the chat head name according user interaction
function pmHim(uname) {
	const name = document.getElementById("chat-user-name");
	name.innerText = uname;
}

function openCreateGroup() {
	const chatPartHeader = document.getElementById("dashboard-chat-part-header");
	const chatPartBody = document.getElementById("dashboard-chat-part-body");
	const createGroupHeader = document.getElementById("dashboard-create-group-header");
	const createGroupBody = document.getElementById("dashboard-create-group-body");
	const welcomeChatSec = document.getElementById("welcome-chat-sec");

	console.log("openCreateGroup function");
	if(welcomeChatSec.classList.contains('d-flex')) {
		console.log("inside if condition");
		welcomeChatSec.classList.remove('d-flex');
		welcomeChatSec.classList.add('d-none');
	} else {
		console.log("inside else condition");
		chatPartHeader.classList.add("d-none");
		chatPartBody.classList.add("d-none");
	}
	
	createGroupHeader.classList.remove("d-none");
	createGroupBody.classList.remove("d-none");

	createGroupHeader.classList.add("d-block");
	createGroupBody.classList.add("d-block");
}

function closeCreateGroup() {
	const chatPartHeader = document.getElementById("dashboard-chat-part-header");
	const chatPartBody = document.getElementById("dashboard-chat-part-body");
	const createGroupHeader = document.getElementById("dashboard-create-group-header");
	const createGroupBody = document.getElementById("dashboard-create-group-body");
	const welcomeChatSec = document.getElementById("welcome-chat-sec");

	if(welcomeChatSec.classList.contains('d-none')) {
		welcomeChatSec.classList.remove('d-none');
		welcomeChatSec.classList.add('d-flex');
	}
	
	createGroupHeader.classList.add("d-none");
	createGroupBody.classList.add("d-none");
}

// Upload image to the create group 
$("#profileImage").click(function(e) {
    $("#imageUpload").click();
});

function fasterPreview( uploader ) {
    if ( uploader.files && uploader.files[0] ){
          $('#profileImage').attr('src', 
             window.URL.createObjectURL(uploader.files[0]) );
    }
}

$("#imageUpload").change(function(){
    fasterPreview( this );
});

jQuery('#createNewGroupForm').validate({
	rules: {
	  groupName: {
	  	required: true,
	  },
	}, messages: {
	  groupName: 'Please enter a name for the group.',
	}, submitHandler: function (createNewGroupForm) {
	  createNewGroupForm.submit();
	}
});
