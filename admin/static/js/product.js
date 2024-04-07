var variants;
var attributes;

function addSpecification(event){
    event.preventDefault()
    const specificationContainer = document.getElementById("specifications")
    const newSpecification=document.createElement("span");
    newSpecification.id="specifications_"+(specificationContainer.childElementCount+1)
    newSpecification.className = "row-padder"
    const input1=document.createElement("input")
    input1.type="text"
    input1.className="section-input"
    input1.placeholder="Height"
    // input1.id=`specifications[${specificationContainer.childElementCount+1}].key`
    // input1.name=`specifications[${specificationContainer.childElementCount+1}].key`
    newSpecification.appendChild(input1);

    const input2=document.createElement("input")
    input2.type="text"
    input2.className="section-input"
    input2.placeholder="1m"
    // input2.id=`specifications[${specificationContainer.childElementCount+1}].value`
    // input2.name=`specifications[${specificationContainer.childElementCount+1}].value`
    newSpecification.appendChild(input2);
    specificationContainer.appendChild(newSpecification)
}

function ShippingCheckBoxHandling(){
    const isPhysicalProduct = document.getElementById('physical_product');
    isPhysicalProduct.addEventListener('change', e => {
        if(e.target.checked === true) {
           document.getElementById('shipping_weight').style.display = 'block';
        }
        if(e.target.checked === false) {
            document.getElementById('shipping_weight').style.display = 'none';
        }
    });
}

async function createProduct(event){
    event.preventDefault();

    const requestBody = await processProductData(event.target)
    const response = await fetch("/product",{method:"POST",body:JSON.stringify(requestBody)})

    if(response.status === 200){
        window.location.assign("/")
        return
    }

}

async function createProduct2(event,product){
    event.preventDefault();
    const formData = new FormData(event.target);
    const formProps = Object.fromEntries(formData);
    var finalProduct={...product,...formProps}
    finalProduct['attributes']=attributes;
    finalProduct["description"]=tinymce.get("description").getContent();
    const specs = document.getElementById("specifications")
    const childs = specs.childElementCount
    let specifications = []
    for(let i=0; i<childs; i++){
        let new_key_value_pair = {}
        new_key_value_pair.key = specs.children[i].children[0].value
        new_key_value_pair.value = specs.children[i].children[1].value

        specifications.push(new_key_value_pair)
    }
    finalProduct["specifications"] = specifications
    finalProduct.price=variants[0].price
    var lowestPriceIndex=0
    for(var i=0;i<variants.length;++i){
        variants[i].images=await uploadImages(variants[i].images)
        if(variants[i].images.length>0){
            variants[i].thumbnail=variants[i].images[0]
        }else{
            variants[i].thumbnail=""
        }
        if(variants[i].price.sellingPrice<finalProduct.price.sellingPrice){
            lowestPriceIndex=i
        }
        delete variants[i].id
    }

    finalProduct.price=variants[lowestPriceIndex].price
    finalProduct.images=variants[lowestPriceIndex].images
    finalProduct.thumbnail=variants[lowestPriceIndex].thumbnail

    console.log("these are values:",finalProduct,formProps,variants);
    var response = await fetch("/product",{method:"POST",body:JSON.stringify(finalProduct)})
    if(response.status !== 200){
        console.log("Unable to edit product")
    }
    const res=await response.json();
    for(var i=0;i<variants.length;++i){
        variants[i].productID=res.id
    }
    response=await fetch("/batch-add-variants",{method:"POST",body:JSON.stringify(variants)})
    if(response.status === 200){
        window.location.assign("/products")
        return
    }
}

async function deleteProduct(event,productID){
    event.preventDefault();

    const response = await fetch("/product?id="+productID,{method:"DELETE"})

    if(response.status === 200){
        window.location.assign("/")
        return
    }

}

async function duplicateProduct(event,productID){
    event.preventDefault();

    const response = await fetch("/product?id="+productID,{method:"PATCH"})
    const respnseData = await response.json()
    if(response.status === 200){
        window.location.assign("/product/"+respnseData.id)
        return
    }

}

async function editProduct(event,productID,prevImages){
    event.preventDefault();

    let requestBody = await processProductData(event.target,prevImages)
    requestBody["id"] = productID
    const response = await fetch("/product?id="+productID,{method:"PUT",body:JSON.stringify(requestBody)})

    if(response.status === 200){
        window.location.assign("/")
        return
    }
}

async function editProduct2(event,product,originalVariants){
    event.preventDefault();
    const formData = new FormData(event.target);
    const formProps = Object.fromEntries(formData);
    var finalProduct={...product,...formProps}
    finalProduct['attributes']=attributes;
    finalProduct["description"]=tinymce.get("description").getContent();
    const specs = document.getElementById("specifications")
    const childs = specs.childElementCount
    let specifications = []
    for(let i=0; i<childs; i++){
        let new_key_value_pair = {}
        new_key_value_pair.key = specs.children[i].children[0].value
        new_key_value_pair.value = specs.children[i].children[1].value

        specifications.push(new_key_value_pair)
    }
    finalProduct["specifications"] = specifications
    const productID=finalProduct["id"]
    console.log("these are values:",finalProduct,formProps,variants);
    finalProduct.price=variants[0].price
    var lowestPriceIndex=0
    for(var i=0;i<variants.length;++i){
        variants[i].images=await uploadImages(variants[i].images)
        if(variants[i].price.sellingPrice<finalProduct.price.sellingPrice){
            lowestPriceIndex=i
        }
        variants[i].thumbnail=variants[i].images[0]
        delete variants[i].id
    }

    finalProduct.price=variants[lowestPriceIndex].price
    finalProduct.images=variants[lowestPriceIndex].images
    finalProduct.thumbnail=variants[lowestPriceIndex].thumbnail
    
    var response = await fetch("/product?id="+productID,{method:"PUT",body:JSON.stringify(finalProduct)})
    if(response.status !== 200){
        console.log("Unable to edit product")
    }
    response=await fetch("/batch-delete-variants?productId="+productID,{method:"DELETE"})
    response=await fetch("/batch-add-variants",{method:"POST",body:JSON.stringify(variants)})
    if(response.status === 200){
        window.location.assign("/products")
        return
    }
}

async function uploadImages(uploadList){
    const imageArr=[]
    if(Array.isArray(uploadList)){
        return uploadList
    }
    else{
        for(var i=0;i<uploadList.length;++i){
            const requestData = new FormData();
            requestData.append('image',  uploadList[i]);
            const response = await fetch('/upload-image', {
                method: 'POST',
                body: requestData,
                headers: {
                    'Accept': 'application/json'
                },
            })
            const responseData = await response.json()
            imageArr.push(responseData.imageURL)
        }
        return imageArr
    }
}

async function processProductData(eventData,prevImages = []){
    const formData = new FormData(eventData);
    const formProps = Object.fromEntries(formData);
    console.log("Data:",formProps,prevImages)
    
    const imagesData =  document.getElementById("images");
    let images = [];

    try{
        if(imagesData.files.length>0){

            const gallery = document.getElementById("gallery")
            const imagesOrder = []
            for(let i=0; i<gallery.childElementCount; i++){
                imagesOrder.push(Number(gallery.children[i].children[0].id))
            }

            for (var i = 0; i < imagesData.files.length; i++){
                const requestData = new FormData();
                requestData.append('image',  imagesData.files[imagesOrder[i]-1]);
                
                const response = await fetch('/upload-image', {
                    method: 'POST',
                    body: requestData,
                    headers: {
                        'Accept': 'application/json'
                    },
                })
    
                const responseData = await response.json()
                // images.push(responseData.imageURL+"."+imagesData.files[i].type.split("/")[imagesData.files[i].type.split("/").length-1])
                images.push(responseData.imageURL)
            }
        }else{
            images = prevImages
        }
    }catch(e){
        console.error("Error occured while uplading image:", e)
    }

    formProps.images = images;
    formProps.thumbnail = images[0];
    
    const price = {};
    price.sellingPrice = Number(formProps.price)
    price.mrp = Number(formProps.mrp)
    price.discount = Number(price.mrp-price.sellingPrice)
    price.discountPercentage = Number(price.discount/price.mrp)*100
    
    formProps.price = price
    
    const specs = document.getElementById("specifications")
    const childs = specs.childElementCount

    let specifications = []
    for(let i=0; i<childs; i++){
        let new_key_value_pair = {}
        new_key_value_pair.key = specs.children[i].children[0].value
        new_key_value_pair.value = specs.children[i].children[1].value

        specifications.push(new_key_value_pair)
    }
    formProps["specifications"] = specifications
    formProps["quantity"] = Number(formProps["quantity"])
    formProps["description"]=tinymce.get("description").getContent();

    delete formProps["mrp"]
    return formProps
}

function PreviewMultipleImages(input){

    placeToInsertImagePreview = document.getElementById("gallery")
    placeToInsertImagePreview.replaceChildren()

    // Multiple images preview in browser
    if (input.files) {
        var filesAmount = input.files.length;

        for (i = 0; i < filesAmount; i++) {

            const src = window.URL.createObjectURL(input.files[i]);

            const imageContainer = document.createElement("div")
            imageContainer.className = "media_image_container"
            imageContainer.draggable="true" 
            imageContainer.ondragstart=function() { drag(event) }; 
            imageContainer.ondrop=function() { drop(event) }; 
            imageContainer.ondragover=function() { allowDrop(event) };
            
            const imgTag = document.createElement("img")
            imgTag.className = "media_image"
            imgTag.id = (placeToInsertImagePreview.childElementCount+1).toString()
            imgTag.src = src

            imageContainer.appendChild(imgTag)
            placeToInsertImagePreview.appendChild(imageContainer)
        }
    }
}

function allowDrop(ev) {
    ev.preventDefault();
  }
  
function drag(ev) {
    ev.dataTransfer.setData("media_image", ev.target.id);
}
  
function drop(ev) {
    ev.preventDefault();

    const parent = ev.target.parentNode.parentNode
    const dragged = ev.target.parentNode
    var data = ev.dataTransfer.getData("media_image");
    var movedElement = document.getElementById(data).parentNode
    
    let currentElement = dragged
    let isRightSibling = false

    while(currentElement.nextElementSibling){
        currentElement = currentElement.nextElementSibling
        if (currentElement.id === movedElement.id){
            isRightSibling = true;
            break;
        }
    }

    if(isRightSibling){
        parent.insertBefore(movedElement,dragged)
    }else{
        parent.insertBefore(movedElement,dragged.nextSibling)
    }
}

function addAttributeValue(event,index){
    event.preventDefault();
    if(isEmpty(index)){
        return;
    }
    const temp=attributes[index]["options"].length
    const attributeContainer = document.getElementById("attribute-options_"+index)
    const newAttributeValue=document.createElement("div");
    newAttributeValue.className="attribute-option-item";
    const input=document.createElement("input");
    input.type="text"
    input.className="section-input option-input"
    input.placeholder="Value"
    input.oninput=function(e){
        e.preventDefault();
        optionValueChange(e,index,temp)
    }
    newAttributeValue.appendChild(input);
    attributeContainer.appendChild(newAttributeValue)

    var tempAttri=attributes
    tempAttri[index].options.push("");
    attributes=tempAttri
    addVariants(tempAttri[index]);
}

function isEmpty(index){
    console.log("isEmpty:",attributes,index)
    return attributes[index]["name"]==''
}

function addVariants(attribute){
    var temp=[...variants];
    var arr=[]
    temp.forEach((item,index)=>{
        console.log("These are options:",attribute["options"])
        if(attribute["options"][0]==''){
            console.log("In this case");
        }else{
        if(item.attributes[attribute["name"]]==attribute["options"][0]){
            var tempVar = JSON.parse(JSON.stringify(item));
            tempVar["attributes"][attribute["name"]]=""
            arr.push(tempVar)
            console.log("tempVar2:",tempVar)
        }}
    })
    var tempVariants=[...temp,...arr];
    // tempVariants=cleanVariants(tempVariants)
    variants=tempVariants;
    console.log("addVariants end:",tempVariants)
}   

function addAttributeItem(event){
    event.preventDefault()
    
    const attributeListContainer = document.getElementById("attribute-list")
    const attributeCount=attributeListContainer.children.length;

    const newAttributeItem=document.createElement("div");
    newAttributeItem.className="attribute-item";

    const optionNameLabel=document.createElement("label");
    optionNameLabel.innerText="Option Name"
    optionNameLabel.className="section-label"
    newAttributeItem.appendChild(optionNameLabel);

    const input=document.createElement("input");
    input.type="text"
    input.className="section-input"
    input.placeholder="Name"
    input.oninput=function(e){ e.preventDefault();optionNameChange(e,attributeCount);}
    newAttributeItem.appendChild(input);

    const optionValueLabel=document.createElement("label");
    optionValueLabel.innerText="Option Values"
    optionValueLabel.className="section-label"
    newAttributeItem.appendChild(optionValueLabel);

    const attributeOptions=document.createElement("div");
    attributeOptions.id="attribute-options_"+(attributeCount)
    newAttributeItem.appendChild(attributeOptions);

    const addAttributeButton=document.createElement("button");
    addAttributeButton.className="add-option"
    addAttributeButton.innerText="+ Add Option"
    addAttributeButton.onclick=function(event){
        event.preventDefault();
        addAttributeValue(event,attributeCount);
    }
    newAttributeItem.appendChild(addAttributeButton);
    attributeListContainer.appendChild(newAttributeItem)
    attributes=[...attributes,{name:"",visType:0,options:[]}]
    var temp=variants;
    for(var i=0;i<temp.length;++i){
        temp[i].attributes[""]=""
    }
    variants=temp;
    console.log("variants:",temp)
}

function setAttributes(attributeList,variantList){
    attributes=attributeList;
    variants=variantList;
    // if(variantList.length==1 && variantList[0].attributes=={}){
    //     emptyVariant=variantList[0]
    //     setEmptyValues();
    // }
    console.log(attributeList,variantList)
    makeVariantArrFromAttributes();
}

// function fillValue(){
//     var attributeArr=[]
//     for(var i=0;i<attributes.length;++i){
//         attributeArr.push(attributes[i].name)
//     }
//     for(var i=0;i<variants.length;++i){
//         var strProperty=[];
//         for(var j=0;j<attributeArr.length;++j){
//             strProperty.push(variants[i].attributes[attributeArr[j]]);
//         }
//         fillInElement(variants[i],strProperty.join("-"),i);
//     }
// }

function fillValue(){
    var attributeArr=[]
    for(var i=0;i<attributes.length;++i){
        attributeArr.push(attributes[i].name)
    }
    for(var i=0;i<variants.length;++i){
        var strProperty=[];
        for(var j=0;j<attributeArr.length;++j){
            strProperty.push(variants[i].attributes[attributeArr[j]]);
        }
        fillInElement(variants[i],strProperty.join("-"),i);
    }
}

function fillInElement(variant,keyString,index){
    document.getElementById('images_'+keyString).oninput=function(event){
        event.preventDefault();
        console.log("Updating images:",event.target.files)
        variants[index].images=event.target.files;
        console.log("variants:",variants)
    }
    document.getElementById('price_'+keyString).value=variant.price.sellingPrice
    document.getElementById('price_'+keyString).oninput=function(event){
        event.preventDefault();
        console.log("Updating sellingPrice:",event.target.value)
        variants[index].price.sellingPrice=Number(event.target.value);
        variants[index].price.discount=Number(variants[index].price.mrp-variants[index].price.sellingPrice);
        variants[index].price.discountPercentage=Number(variants[index].price.discount/variants[index].price.mrp)*100;
        console.log("variants:",variants)
    }
    document.getElementById('mrp_'+keyString).value=variant.price.mrp
    document.getElementById('mrp_'+keyString).oninput=function(event){
        event.preventDefault();
        console.log("Updating mrp:",event.target.value)
        variants[index].price.mrp=Number(event.target.value);
        variants[index].price.discount=Number(variants[index].price.mrp-variants[index].price.sellingPrice);
        variants[index].price.discountPercentage=Number(variants[index].price.discount/variants[index].price.mrp)*100;
        console.log("variants:",variants)
    }
    document.getElementById('quantity_'+keyString).value=variant.quantity
    document.getElementById('quantity_'+keyString).oninput=function(event){
        event.preventDefault();
        console.log("Updating quantity:",event.target.value)
        variants[index].quantity=Number(event.target.value);
        console.log("variants:",variants)
    }
    document.getElementById('sku_'+keyString).value=variant.sku
    document.getElementById('sku_'+keyString).oninput=function(event){
        event.preventDefault();
        console.log("Updating sku:",event.target.value)
        variants[index].sku=event.target.value;
        console.log("variants:",variants)
    }
    document.getElementById('barcode_'+keyString).value=variant.barcode
    document.getElementById('barcode_'+keyString).oninput=function(event){
        event.preventDefault();
        console.log("Updating barcode:",event.target.value)
        variants[index].barcode=event.target.value;
        console.log("variants:",variants)
    }
    document.getElementById('weight_'+keyString).value=variant.weight
    document.getElementById('weight_'+keyString).oninput=function(event){
        event.preventDefault();
        console.log("Updating weight:",event.target.value)
        variants[index].weight=event.target.value;
        console.log("variants:",variants)
    }
    document.getElementById('length_'+keyString).value=variant.length
    document.getElementById('length_'+keyString).oninput=function(event){
        event.preventDefault();
        console.log("Updating length:",event.target.value)
        variants[index].length=event.target.value;
        console.log("variants:",variants)
    }
    document.getElementById('breadth_'+keyString).value=variant.breadth
    document.getElementById('breadth_'+keyString).oninput=function(event){
        event.preventDefault();
        console.log("Updating breadth:",event.target.value)
        variants[index].breadth=event.target.value;
        console.log("variants:",variants)
    }
    document.getElementById('height_'+keyString).value=variant.height
    document.getElementById('height_'+keyString).oninput=function(event){
        event.preventDefault();
        console.log("Updating height:",event.target.value)
        variants[index].height=event.target.value;
        console.log("variants:",variants)
    }
}

function makeVariantArrFromAttributes(){
    const variantCover=document.getElementById('variant-cover');
    var temp=[]
    for(var i=0;i<attributes.length;++i){
        temp.push(attributes[i].options)
    }
    if(temp.length==0){
        variantCover.replaceChildren(variantElement("",0))
        fillValue();
    }else{
        var cartesianProduct=cartesian(...temp)
        var arr=[]
        for(var i=0;i<cartesianProduct.length;++i){
            arr.push(variantElement(cartesianProduct[i],i));
        }
        variantCover.replaceChildren(...arr)
        fillValue();
    }
}

// function getVariant(variants,nameArr,optionsArr){
//     var flag=0;
//     for(var i=0;i<variants.length;++i){
//         if(variants[i].attributes[nameArr[0]]==optionsArr[0]){
//             flag=1;
//             for(var j=1;j<optionsArr.length;++j){
//                 if(variants[i].attributes[nameArr[j]]!==optionsArr[j]){
//                     flag=0;
//                     break;
//                 }
//             }
//             if(flag){
//                 return variants[i];
//             }
//         }
//     }
//     return {};
// }

function getVariant(variants,nameArr,optionsArr){
    var flag=0;
    for(var i=0;i<variants.length;++i){
        if(variants[i].attributes[nameArr[0]]==optionsArr[0]){
            flag=1;
            for(var j=1;j<optionsArr.length;++j){
                if(variants[i].attributes[nameArr[j]]!==optionsArr[j]){
                    flag=0;
                    break;
                }
            }
            if(flag){
                return variants[i];
            }
        }
    }
    return {};
}

const cartesian = (...a) => a.reduce((a, b) => a.flatMap(d => b.map(e => [d, e].flat())));

function variantElement(optionArr,index){
    const returnElement=document.createElement("div");
    returnElement.className="variant-item"
    returnElement.id="variant_"+index

    var slug=optionArr
    var mainSlug=optionArr

    if(Array.isArray(optionArr)){
        slug=optionArr.join("-")
        mainSlug=optionArr.join("/")
    }

    returnElement.innerHTML=`
        <div class="variant-input">
            <div>`+mainSlug+`</div>
        </div>
        <div class="variant-input">
            <input type="file" class="images" multiple placeholder="Images" id="images_`+slug+`" for="images_`+slug+`"/>
        </div>
        <div class="variant-input">
            <input type="number" placeholder="Price" id="price_`+slug+`" for="price_`+slug+`"/>
        </div>
        <div class="variant-input">
            <input type="number" placeholder="MRP" id="mrp_`+slug+`" for="mrp_`+slug+`"/>
        </div>
        <div class="variant-input">
            <input type="number" placeholder="Stock" id="quantity_`+slug+`" for="quantity_`+slug+`" />
        </div>
        <div class="variant-input">
            <input placeholder="SKU" id="sku_`+slug+`" for="sku_`+slug+`"/>
        </div>
        <div class="variant-input">
            <input placeholder="Barcode" id="barcode_`+slug+`" for="barcode_`+slug+`"/>
        </div>
        <div class="variant-input">
            <input type="number" step="0.01" placeholder="Weight (in kgs)" id="weight_`+slug+`" for="weight_`+slug+`"/>
        </div>
        <div class="variant-input">
            <input type="number" step="0.01" placeholder="Length (in cms)" id="length_`+slug+`" for="length_`+slug+`"/>
        </div>
        <div class="variant-input">
            <input type="number" step="0.01" placeholder="Breadth (in cms)" id="breadth_`+slug+`" for="breadth_`+slug+`"/>
        </div>
        <div class="variant-input">
            <input type="number" step="0.01" placeholder="Height" id="height_`+slug+`" for="height_`+slug+`"/>
        </div>
    `
    return returnElement
}

function optionNameChange(event,index){
    console.log("changing name:",index,attributes[index]["name"],event.target.value)
    updateVariantKey(attributes[index]["name"],event.target.value);
    attributes[index]["name"]=event.target.value;
    if(event.target.value=="Color" || event.target.value=="Colour"){
        attributes[index]["visType"]=1;
    }
}

function optionValueChange(event,attrIndex,valIndex){
    updateVariantValue(attributes[attrIndex]["name"],attributes[attrIndex]["options"][valIndex],event.target.value)
    attributes[attrIndex]["options"][valIndex]=event.target.value;
    makeVariantArrFromAttributes();
}

function updateVariantKey(originalKey,afterKey){
    var temp=variants
    for(var i=0;i<temp.length;++i){
        temp[i].attributes[afterKey]=temp[i].attributes[originalKey]
        delete temp[i].attributes[originalKey]
    }
    // temp=cleanVariants(temp)
    variants=temp
    console.log("updateVariantKey variants:",temp,attributes)
}

function updateVariantValue(key,originalValue,targetValue){
    var temp=variants
    if(originalValue==undefined){
        originalValue=""
    }
    for(var i=0;i<temp.length;++i){
        if(temp[i].attributes[key]==originalValue){
            temp[i].attributes[key]=targetValue
        }
    }
    variants=temp
    console.log("updateVariantValue variants:",temp)
}

