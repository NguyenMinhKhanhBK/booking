let attention = Prompt();

(function() {
    'use strict';
    window.addEventListener('load', function() {
        // Fetch all the forms we want to apply custom Bootstrap validation styles to
        let forms = document.getElementsByClassName('needs-validation');
        // Loop over them and prevent submission
        Array.prototype.filter.call(forms, function(form) {
            form.addEventListener('submit', function(event) {
                if (form.checkValidity() === false) {
                    event.preventDefault();
                    event.stopPropagation();
                }
                form.classList.add('was-validated');
            }, false);
        });
    }, false);
})();


function notify(msg, msgType) {
    notie.alert({
        type: msgType,
        text: msg,
    })
}

function notifyModal(title, text, icon, confirmationButtonText) {
    Swal.fire({
        title: title,
        html: text,
        icon: icon,
        confirmButtonText: confirmationButtonText
    })
}


function Prompt() {
    let toast = function(c) {
        const {
            msg = '',
                icon = 'success',
                position = 'top-end',

        } = c;

        const Toast = Swal.mixin({
            toast: true,
            title: msg,
            position: position,
            icon: icon,
            showConfirmButton: false,
            timer: 3000,
            timerProgressBar: true,
            didOpen: (toast) => {
                toast.addEventListener('mouseenter', Swal.stopTimer)
                toast.addEventListener('mouseleave', Swal.resumeTimer)
            }
        })

        Toast.fire({})
    }

    let success = function(c) {
        const {
            msg = "",
                title = "",
                footer = "",
        } = c

        Swal.fire({
            icon: 'success',
            title: title,
            text: msg,
            footer: footer,
        })

    }

    let error = function(c) {
        const {
            msg = "",
                title = "",
                footer = "",
        } = c

        Swal.fire({
            icon: 'error',
            title: title,
            text: msg,
            footer: footer,
        })

    }

    async function custom(c) {
        const {
            icon = "",
                msg = "",
                title = "",
                showConfirmButton = true,
        } = c;

        const {
            value: result
        } = await Swal.fire({
            icon: icon,
            title: title,
            html: msg,
            backdrop: false,
            allowOutsideClick: false,
            focusConfirm: false,
            showCancelButton: true,
            showConfirmButton: showConfirmButton,
            willOpen: () => {
                if (c.willOpen !== undefined) {
                    c.willOpen();
                }
            },
            didOpen: () => {
                if (c.didOpen !== undefined) {
                    c.didOpen();
                }
            },
        })

        if (result) {
            if (result.dismiss !== Swal.DismissReason.cancel) {
                if (result.value !== "") {
                    if (c.callback !== undefined) {
                        c.callback(result);
                    }
                } else {
                    c.callback(false);
                }
            } else { // users press CANCEL button
                c.callback(false);
            }

        }


    }

    return {
        toast: toast,
        success: success,
        error: error,
        custom: custom,
    }
}

function buttonHandler(roomID, csrfToken) {
        let html = `
        <form id="check-availability-form" action="/search-availability-json" method="post" novalidate class="needs-validation" autocomplete="off">
            <div class="row">
                <div class="col">
                    <div class="row" id="reservation-dates-modal">
                        <div class="col">
                            <input disabled required class="form-control" type="text" name="start" id="start" placeholder="Arrival">
                        </div>
                        <div class="col">
                            <input disabled required class="form-control" type="text" name="end" id="end" placeholder="Departure">
                        </div>

                    </div>
                </div>
            </div>
        </form>
        `;
        attention.custom({
            title: 'Choose your dates',
            msg: html,
            willOpen: () => {
                const elem = document.getElementById("reservation-dates-modal");
                const rp = new DateRangePicker(elem, {
                    format: 'yyyy-mm-dd',
                    showOnFocus: true,
                    minDate: new Date(),
                })
            },
            didOpen: () => {
                document.getElementById("start").removeAttribute("disabled");
                document.getElementById("end").removeAttribute("disabled");
            },
            callback: function(result) {
                let form = document.getElementById("check-availability-form");
                let formData = new FormData(form);
                formData.append("csrf_token", csrfToken);
                formData.append("room_id", roomID);

                    fetch("/search-availability-json", {
                            method: "post",
                            body: formData,
                        })
                    .then(rsp => rsp.json())
                    .then(data => {
                            if (data.ok) {
                                attention.custom({
                                    icon: 'success',
                                    msg: '<p>Room is available!</p>'
                                        + '<p><a href="/book-room?id='
                                        + data.room_id
                                        + '&s='
                                        + data.start_date
                                        + '&e='
                                        + data.end_date
                                        + '" class="btn btn-primary">'
                                        + 'Book now!</a></p>',

                                    showConfirmButton: false,
                                })
                            } else {
                                attention.error({
                                    msg: "No availability",
                                });
                            }
                    })
            }
        });

}
