# HTMLnoJS Demo API Handlers

def htmx_hello(request):
    """
    Simple hello world HTMX handler
    Returns a friendly greeting message
    """
    return '<div class="alert alert-success">🎉 Hello from HTMLnoJS! HTMX is working perfectly!</div>'

def htmx_form(request):
    """
    Handle form submission demo
    Processes form data and returns formatted response
    """
    message = request.get('message', 'No message provided')
    return f'''
    <div class="alert alert-info">
        <h3>Form Submitted Successfully!</h3>
        <p><strong>Your message:</strong> {message}</p>
        <small>Processed by Python backend via HTMX</small>
    </div>
    '''