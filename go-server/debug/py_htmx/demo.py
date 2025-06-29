# demo.py - Debug version
def htmx_hello(request):
    """Simple hello world HTMX handler"""
    return '''
    <div class="alert alert-success">
        <strong>Hello from HTMLnoJS!</strong><br>
        HTMX is working perfectly! 🎉
    </div>
    '''

def htmx_form(request):
    """Handle form submission demo - DEBUG VERSION"""
    print(f"DEBUG: Received request: {request}")
    print(f"DEBUG: Request type: {type(request)}")

    if hasattr(request, 'keys'):
        print(f"DEBUG: Request keys: {list(request.keys())}")
        for key in request.keys():
            print(f"DEBUG: {key} = {request[key]}")
    elif isinstance(request, dict):
        print(f"DEBUG: Request dict: {request}")
    else:
        print(f"DEBUG: Request attributes: {dir(request)}")

    # Try different ways to get the message
    message = None

    # Method 1: dict access
    if isinstance(request, dict):
        message = request.get('message')
        print(f"DEBUG: Dict access message: {message}")

    # Method 2: attribute access
    if hasattr(request, 'message'):
        message = getattr(request, 'message')
        print(f"DEBUG: Attribute access message: {message}")

    # Method 3: form data
    if hasattr(request, 'form'):
        print(f"DEBUG: Form data: {request.form}")
        message = request.form.get('message')
        print(f"DEBUG: Form access message: {message}")

    # Fallback
    if not message:
        message = 'No message provided (DEBUG MODE)'

    return f'''
    <div class="alert alert-info">
        <strong>Form Submitted Successfully!</strong><br>
        Your message: "{message}"<br>
        <small>Processed by Python backend via HTMX</small><br>
        <small>Debug info: Request type was {type(request)}</small>
    </div>
    '''