
function addUrl(elem){
if(elem!=null){
    elem.innerHTML=document.location.hostname+
    ((document.location.port!="80")?":"+document.location.port:"")+document.location.pathname
}
}
addUrl(document.getElementById("url"))

function fillData(){
    var div = document.getElementById("data")
    var elem = document.createElement("a")
    elem.textContent="Data"
    elem.href=document.location.pathname+"?view=json"
    div.appendChild(elem)
    /*fetch(document.location.pathname+"?view=json")
    .then(function(response) {
        return response.json();
      })
      .then(function(json) {
        
        div.innerText = json
       
        
      });*/
}
fillData()