const React = window.React;
const ReactDOM = window.ReactDOM;



function addUrl(elem){
if(elem!=null){
    elem.innerHTML=document.location.hostname+
    ((document.location.port.length!=0)?":"+document.location.port:"")+document.location.pathname
}
}
addUrl(document.getElementById("url"))

/*ReactDOM.render(
    <a href={document.location.pathname+"?view=json"}>
    Par ici le data</a>,
    document.getElementById("data")
);*/

function fillData(){
    var div = document.getElementById("accordion")
    var divdatalink = document.getElementById("data")
    

    divdatalink.href=document.location.pathname+"?view=json";
    fetch(document.location.pathname+"?view=json")
    .then(function(response) {
        return response.json();
      })
      .then(function(data) {
          data = data.reverse(); // but the last query at the top
        for (var i = 0; i < data.length; i++) {
            var listItem = document.createElement('li');
            var d = new Date(Date.parse(data[i].Time));
            listItem.innerHTML = '<strong>' + data[i].Method + '</strong> From ' +
                                 data[i].RemoteAddr +
                                 ' at <strong>' + d.getHours()+'h'+d.getMinutes()  + '</strong>';

            var accor = document.createElement("div");
            accor.className="card";
            accor.innerHTML='<div class="card-header" id="headingOne">'+
                            '<h5 class="mb-0">'+
                            '<button class="btn btn-link" type="button" data-toggle="collapse" data-target="#collapse'+i+'" aria-expanded="true" aria-controls="collapse'+i+'">'+
                            '<strong>'+data[i].Method+'</strong> '+ data[i].RemoteAddr+' '+ d.getHours()+'h'+d.getMinutes()  +
                            '</button>'+
                            '</h5>'+
                            '</div>'+
                            '<div id="collapse'+i+'" class="collapse '+(i==0?'show':'')+'" aria-labelledby="headingOne" data-parent="#accordion">'+
                            '<div class="card-body">'+
                            '<ul class="list-group list-group-flush">'+
                            '<li class="list-group-item"><strong>Method:</strong> '+data[i].Method+'</li>'+
                            '<li class="list-group-item"><strong>User-Agent:</strong> '+data[i].Header["User-Agent"]+'</li>'+
                            '<li class="list-group-item"><strong>Query:</strong> '+data[i].URL["RawQuery"]+'</li>'+
                            '<li class="list-group-item"><strong>Body:</strong> '+'<pre><code>'+data[i].Body+'</code></pre>'+'</li>'+
                            //'<li class="list-group-item"><strong>'++'</strong> '++'</li>'+
                            '</ul>'+
                            
                            
                            
                            '</div>'+
                            '</div>'+
                            '</div>';
            div.appendChild(accor);
          }
        //div.innerText = json
       
        
      });
}
fillData()